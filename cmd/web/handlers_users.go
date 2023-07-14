package main

import (
	"domainator/internal/config"
	"domainator/internal/services"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) signupForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "signup.html.tmpl", &map[string]any{"Year": time.Now().Year()})
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	var payload services.UserCredentials
	app.decodeForm(r, &payload)

	ok := app.userSvc.Validate(&payload)
	if !ok {
		templateData := map[string]any{
			"Year": time.Now().Year(),
			"Form": payload,
		}
		app.render(w, http.StatusOK, "signup.html.tmpl", &templateData)
		return
	}

	u, err := app.userSvc.New(payload.Email, payload.Password)
	if err != nil {
		app.serverError(w, err)
		return
	}
	_, err = app.userSvc.Create(r.Context(), u)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateEmail) {
			templateData := map[string]any{
				"Year":  time.Now().Year(),
				"Form":  payload,
				"Flash": "Email already in use",
			}
			app.render(w, http.StatusOK, "signup.html.tmpl", &templateData)
		} else {
			app.serverError(w, err)
		}
		return
	}

	// TODO: redirect to a page that says "check your email for activation link"
	// and actually implement the activation part.
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *application) loginForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "login.html.tmpl", &map[string]any{"Year": time.Now().Year()})
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	var payload services.UserCredentials
	app.decodeForm(r, &payload)

	ok := app.userSvc.Validate(&payload)
	if !ok {
		templateData := map[string]any{
			"Year": time.Now().Year(),
			"Form": payload,
		}
		app.render(w, http.StatusOK, "login.html.tmpl", &templateData)
		return
	}

	u, err := app.userSvc.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		if errors.Is(err, services.ErrNotFound) {
			templateData := map[string]any{
				"Year":  time.Now().Year(),
				"Form":  payload,
				"Flash": "Invalid email or password",
			}
			app.render(w, http.StatusOK, "login.html.tmpl", &templateData)
			return
		}

		app.serverError(w, err)
		return
	}

	match := u.Password.Matches(payload.Password)
	if !match {
		templateData := map[string]any{
			"Year":  time.Now().Year(),
			"Form":  payload,
			"Flash": "Invalid email or password",
		}
		app.render(w, http.StatusOK, "login.html.tmpl", &templateData)
		return
	}

	if !u.Activated {
		templateData := map[string]any{
			"Year":  time.Now().Year(),
			"Form":  payload,
			"Flash": "Please activate your account in order to continue",
		}
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
	app.logit.Info("User logging out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
