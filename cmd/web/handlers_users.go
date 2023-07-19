package main

import (
	"domainator/internal/config"
	"domainator/internal/services"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) signupForm(w http.ResponseWriter, r *http.Request) {
	templateData := initialTmplData(r)
	app.render(w, http.StatusOK, "signup.html.tmpl", &templateData)
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	var payload services.UserCredentials
	app.decodeForm(r, &payload)
	templateData := initialTmplData(r)

	ok := app.userSvc.Validate(&payload)
	if !ok {
		templateData["Form"] = payload
		app.render(w, http.StatusOK, "signup.html.tmpl", &templateData)
		return
	}

	u, err := app.userSvc.New(payload.Email, payload.Password)
	if err != nil {
		app.serverError(w, err)
		return
	}

	_, code, err := app.userSvc.Create(r.Context(), u)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateEmail) {
			templateData["Form"] = payload
			templateData["Flash"] = "Email already in use"
			app.render(w, http.StatusOK, "signup.html.tmpl", &templateData)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// TODO: Email verification code
	app.logit.Info(fmt.Sprintf("Verification code for %s: %s", u.Email, code))
	http.Redirect(w, r, "/user/verify", http.StatusSeeOther)
}

func (app *application) loginForm(w http.ResponseWriter, r *http.Request) {
	templateData := initialTmplData(r)
	app.render(w, http.StatusOK, "login.html.tmpl", &templateData)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var payload services.UserCredentials
	app.decodeForm(r, &payload)
	templateData := initialTmplData(r)

	ok := app.userSvc.Validate(&payload)
	if !ok {
		templateData["Form"] = payload
		app.render(w, http.StatusOK, "login.html.tmpl", &templateData)
		return
	}

	u, err := app.userSvc.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			templateData["Form"] = payload
			templateData["Flash"] = "Invalid email or password"
			app.render(w, http.StatusOK, "login.html.tmpl", &templateData)
			return
		}

		app.serverError(w, err)
		return
	}

	match := u.Password.Matches(payload.Password)
	if !match {
		templateData["Form"] = payload
		templateData["Flash"] = "Invalid email or password"
		app.render(w, http.StatusOK, "login.html.tmpl", &templateData)
		return
	}

	if !u.Activated {
		templateData["Form"] = payload
		templateData["Flash"] = "Please activate your account in order to continue"
		app.render(w, http.StatusOK, "login.html.tmpl", &templateData)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID,
		"exp": time.Now().Add(config.GetDuration("TOKEN_EXP")).Unix(),
		"aud": "domainator",
	})

	t, err := token.SignedString([]byte(config.GetString("JWT_SECRET")))
	if err != nil {
		app.serverError(w, err)
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

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
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

func (app *application) verifyForm(w http.ResponseWriter, r *http.Request) {
	templateData := initialTmplData(r)
	app.render(w, http.StatusOK, "verify.html.tmpl", &templateData)
}

func (app *application) verify(w http.ResponseWriter, r *http.Request) {
	var payload services.VerificationCode
	app.decodeForm(r, &payload)
	templateData := initialTmplData(r)

	ok := app.userSvc.Validate(&payload)
	if !ok {
		templateData["Form"] = payload
		app.render(w, http.StatusOK, "verify.html.tmpl", &templateData)
		return
	}

	err := app.userSvc.Verify(r.Context(), payload.Email, payload.Plain)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCode) {
			templateData["Form"] = payload
			templateData["Flash"] = err.Error()
			app.render(w, http.StatusOK, "verify.html.tmpl", &templateData)
		} else {
			app.serverError(w, err)
		}
		return
	}

	http.Redirect(w, r, "/pings", http.StatusSeeOther)
}
