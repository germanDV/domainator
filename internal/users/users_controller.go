// Package users holds the logic for the users service
package users

import (
	"domainator/internal/config"
	"domainator/internal/httphelp"
	"domainator/internal/logger"
	"domainator/internal/notificators"
	"domainator/internal/notifier"
	"domainator/internal/plans"
	"domainator/internal/tmpl"
	"domainator/internal/validation"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

// Controller is a controller that handles requests to the users service
type Controller struct {
	repo      Repo
	validator *validator.Validate
	mailer    notifier.Notifier
	plansRepo plans.Repo
}

// NewController returns a new users controller
func NewController(repo Repo, validate *validator.Validate, plansRepo plans.Repo) *Controller {
	return &Controller{
		repo:      repo,
		validator: validate,
		mailer:    notifier.NewMailer(),
		plansRepo: plansRepo,
	}
}

// SignupForm renders the page for creating a new user.
func (c *Controller) SignupForm(w http.ResponseWriter, r *http.Request) {
	templateData := tmpl.BaseData(r)
	tmpl.RenderPage(w, http.StatusOK, "signup.html.tmpl", &templateData)
}

// Signup creates a new user, saves it to the database and sends a verification code.
func (c *Controller) Signup(w http.ResponseWriter, r *http.Request) {
	var payload UserCredentials
	httphelp.DecodeForm(r, &payload)
	templateData := tmpl.BaseData(r)

	ok := payload.Validate(c.validator)
	if !ok {
		templateData["Form"] = payload
		tmpl.RenderPage(w, http.StatusOK, "signup.html.tmpl", &templateData)
		return
	}

	u, err := newUser(payload.Email, payload.Password)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	_, code, err := c.repo.Create(r.Context(), u)
	if err != nil {
		if errors.Is(err, validation.ErrDuplicateEmail) {
			templateData["Form"] = payload
			templateData["Flash"] = "Email already in use"
			tmpl.RenderPage(w, http.StatusOK, "signup.html.tmpl", &templateData)
		} else {
			httphelp.ServerError(w, err)
		}
		return
	}

	go func(validationCode string) {
		defer func() {
			if err := recover(); err != nil {
				logger.Writer.Error(fmt.Sprintf("Send verification email panicked: %v", err))
			}
		}()

		sub, body, err := notifier.ParseTemplate("verification.html.tmpl", map[string]any{"Code": validationCode})
		if err != nil {
			logger.Writer.Error(err)
			return
		}

		fmt.Println("This should be an email", sub, body)
		c.mailer.Notify(notifier.Message{
			To:      u.Email,
			Subject: sub,
			Body:    body,
		})
	}(code)

	target := fmt.Sprintf("/user/verify?email=%s", url.QueryEscape(payload.Email))
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// LoginForm renders the page for logging in.
func (c *Controller) LoginForm(w http.ResponseWriter, r *http.Request) {
	templateData := tmpl.BaseData(r)
	email, err := url.QueryUnescape(r.URL.Query().Get("email"))
	if err == nil && email != "" {
		templateData["Form"] = UserCredentials{Email: email}
	}
	tmpl.RenderPage(w, http.StatusOK, "login.html.tmpl", &templateData)
}

// Login logs a user in.
func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var payload UserCredentials
	httphelp.DecodeForm(r, &payload)
	templateData := tmpl.BaseData(r)

	ok := payload.Validate(c.validator)
	if !ok {
		templateData["Form"] = payload
		tmpl.RenderPage(w, http.StatusOK, "login.html.tmpl", &templateData)
		return
	}

	u, err := c.repo.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		if errors.Is(err, validation.ErrNotFound) {
			templateData["Form"] = payload
			templateData["Flash"] = "Invalid email or password"
			tmpl.RenderPage(w, http.StatusOK, "login.html.tmpl", &templateData)
			return
		}

		httphelp.ServerError(w, err)
		return
	}

	match := u.Password.Matches(payload.Password)
	if !match {
		templateData["Form"] = payload
		templateData["Flash"] = "Invalid email or password"
		tmpl.RenderPage(w, http.StatusOK, "login.html.tmpl", &templateData)
		return
	}

	if !u.Activated {
		templateData["Form"] = payload
		templateData["Flash"] = "Please activate your account in order to continue"
		tmpl.RenderPage(w, http.StatusOK, "login.html.tmpl", &templateData)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID,
		"pln": u.PlanID,
		"exp": time.Now().Add(config.GetDuration("TOKEN_EXP")).Unix(),
		"aud": "domainator",
	})

	t, err := token.SignedString([]byte(config.GetString("JWT_SECRET")))
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    t,
		Path:     "/",
		Expires:  time.Now().Add(config.GetDuration("TOKEN_EXP")),
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   true,
	})

	http.Redirect(w, r, "/pings", http.StatusSeeOther)
}

// Logout logs a user out (expires the cookie with the token).
func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
		Secure:   true,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// VerifyForm renders the page for verifying a user's email.
func (c *Controller) VerifyForm(w http.ResponseWriter, r *http.Request) {
	templateData := tmpl.BaseData(r)
	email, err := url.QueryUnescape(r.URL.Query().Get("email"))
	if err == nil && email != "" {
		templateData["Form"] = VerificationCode{Email: email}
	}
	tmpl.RenderPage(w, http.StatusOK, "verify.html.tmpl", &templateData)
}

// Verify verifies a user's email and activates their account.
func (c *Controller) Verify(w http.ResponseWriter, r *http.Request) {
	var payload VerificationCode
	httphelp.DecodeForm(r, &payload)
	templateData := tmpl.BaseData(r)

	ok := payload.Validate(c.validator)
	if !ok {
		templateData["Form"] = payload
		tmpl.RenderPage(w, http.StatusOK, "verify.html.tmpl", &templateData)
		return
	}

	err := c.repo.Verify(r.Context(), payload.Email, payload.Plain)
	if err != nil {
		if errors.Is(err, validation.ErrInvalidCode) {
			templateData["Form"] = payload
			templateData["Flash"] = err.Error()
			tmpl.RenderPage(w, http.StatusOK, "verify.html.tmpl", &templateData)
		} else {
			httphelp.ServerError(w, err)
		}
		return
	}

	target := fmt.Sprintf("/user/login?email=%s", url.QueryEscape(payload.Email))
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// GetSettings renders the page with the user settings.
func (c *Controller) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID := httphelp.GetUserIDFromCtx(w, r)
	user, err := c.repo.GetByID(r.Context(), userID)
	if err != nil {
		httphelp.ClientError(w, http.StatusUnauthorized)
		return
	}

	prefs, err := c.repo.GetNotificationPrefsByUserID(r.Context(), userID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	emailSettings := find[NotificationPref](prefs, func(p NotificationPref) bool {
		return p.Service == notificators.Email
	})
	slackSettings := find[NotificationPref](prefs, func(p NotificationPref) bool {
		return p.Service == notificators.Slack
	})

	plan, err := c.plansRepo.GetByID(r.Context(), user.PlanID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	templateData := tmpl.BaseData(r)
	templateData["User"] = map[string]string{"Email": user.Email}
	templateData["Plan"] = map[string]string{"Name": plan.Name}
	templateData["Prefs"] = map[string]NotificationPref{
		"Email": emailSettings,
		"Slack": slackSettings,
	}

	tmpl.RenderPage(w, http.StatusOK, "settings.html.tmpl", &templateData)
}

// UpsertEmailSetting creates or updates a user's email notification settings.
func (c *Controller) UpsertEmailSetting(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httphelp.ClientError(w, http.StatusBadRequest)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)

	var payload EmailUpdate
	httphelp.DecodeForm(r, &payload)

	ok := payload.Validate(c.validator)
	if !ok {
		data := map[string]any{
			"Prefs": map[string]any{
				"Email": map[string]any{
					"ID": id,
					"To": payload.Email,
				},
			},
			"Error": payload.Errors["Email"],
		}
		tmpl.RenderFragment(w, "settings_email_input.html.tmpl", &data)
		return
	}

	var pref *NotificationPref
	var e error
	if id == 0 {
		pref, err = c.repo.CreateNotification(r.Context(), userID, notificators.Email, payload.Email)
	} else {
		pref, err = c.repo.UpdateNotification(r.Context(), id, userID, payload.Email)
	}
	if e != nil {
		httphelp.ServerError(w, err)
		return
	}

	emailSettings := map[string]any{"ID": pref.ID, "To": pref.To}
	data := map[string]any{
		"Prefs": map[string]any{
			"Email": emailSettings,
		},
	}

	tmpl.RenderFragment(w, "settings_email_input.html.tmpl", &data)
}

// UpsertSlackSetting creates or updates a user's slack notification settings.
func (c *Controller) UpsertSlackSetting(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httphelp.ClientError(w, http.StatusBadRequest)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)

	var payload SlackUpdate
	httphelp.DecodeForm(r, &payload)

	ok := payload.Validate(c.validator)
	if !ok {
		fmt.Printf("%+v\n", payload.Errors)
		data := map[string]any{
			"Prefs": map[string]any{
				"Slack": map[string]any{
					"ID": id,
					"To": payload.Webhook,
				},
			},
			"Error": payload.Errors["Webhook"],
		}
		tmpl.RenderFragment(w, "settings_slack_input.html.tmpl", &data)
		return
	}

	var pref *NotificationPref
	var e error
	if id == 0 {
		pref, err = c.repo.CreateNotification(r.Context(), userID, notificators.Slack, payload.Webhook)
	} else {
		pref, err = c.repo.UpdateNotification(r.Context(), id, userID, payload.Webhook)
	}
	if e != nil {
		httphelp.ServerError(w, err)
		return
	}

	slackSettings := map[string]any{"ID": pref.ID, "To": pref.To}
	data := map[string]any{
		"Prefs": map[string]any{
			"Slack": slackSettings,
		},
	}

	tmpl.RenderFragment(w, "settings_slack_input.html.tmpl", &data)
}

// TogglePref enables/disables a user's notification preference.
func (c *Controller) TogglePref(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httphelp.ClientError(w, http.StatusBadRequest)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	isEnabled, err := c.repo.ToggleNotification(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, validation.ErrNotFound) {
			httphelp.ClientError(w, http.StatusNotFound)
		} else {
			httphelp.ServerError(w, err)
		}
		return
	}

	var html string
	if isEnabled {
		html = fmt.Sprintf(
			`<input class="toggle-checkbox" type="checkbox" name="enabled" hx-put="/settings/toggle/%d" hx-swap="outerHTML" hx-indicator="#ind" checked />`,
			id,
		)
	} else {
		html = fmt.Sprintf(
			`<input class="toggle-checkbox" type="checkbox" name="enabled" hx-put="/settings/toggle/%d" hx-swap="outerHTML" hx-indicator="#ind" />`,
			id,
		)
	}

	w.Write([]byte(html))
}

// find returns the first element in the slice that matches the predicate function.
func find[T any](slice []T, predicate func(T) bool) T {
	for _, v := range slice {
		if predicate(v) {
			return v
		}
	}

	empty := new(T)
	return *empty
}
