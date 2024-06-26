package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/cookies"
	"github.com/germandv/domainator/internal/githubauth"
	"github.com/germandv/domainator/internal/tokenauth"
	"github.com/germandv/domainator/internal/users"
	"golang.org/x/oauth2"
)

const stateCookieName = "the_state_str"

func GithubLogin(logger *slog.Logger, githubConfig *oauth2.Config, cookieSigningSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)
		if userID != "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		state := common.GenerateRandomString(32)

		err := cookies.WriteSigned(w, http.Cookie{
			Name:     stateCookieName,
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			Expires:  time.Now().Add(5 * time.Minute),
		}, cookieSigningSecret)
		if err != nil {
			logger.Error("error writing signed cookie", "err", err.Error())
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		url := githubConfig.AuthCodeURL(state)
		logger.Info("redirecting to GitHub for sign in")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func GithubCallback(
	logger *slog.Logger,
	githubConfig *oauth2.Config,
	authService tokenauth.Service,
	usersService users.Service,
	cookieSigningSecret []byte,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		code := r.FormValue("code")

		stateCookie, err := cookies.ReadSigned(r, stateCookieName, cookieSigningSecret)
		if err != nil || state != stateCookie {
			logger.Error("error reading signed cookie or comparing state", "err", err.Error(), "state", state, "stateCookie", stateCookie)
			http.Error(w, "Invalid or missing state", http.StatusUnauthorized)
			return
		}

		token, err := githubConfig.Exchange(r.Context(), code)
		if err != nil {
			logger.Error("error in GitHub exchange", "err", err.Error())
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		userData, err := githubauth.GetGithubUserData(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if userData.Email == "" {
			email, err := githubauth.GetGithubUserEmail(token)
			if err != nil {
				logger.Error("error getting GitHub user email", "err", err.Error())
				http.Error(w, "something went wrong", http.StatusInternalServerError)
				return
			}
			userData.Email = email
		}

		email, err := users.ParseEmail(userData.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user, err := usersService.GetByEmail(r.Context(), users.GetByEmailReq{Email: email})
		if err != nil {
			if errors.Is(err, users.ErrNotFound) {
				user, err = usersService.Save(r.Context(), users.SaveReq{
					Email:              email,
					Name:               userData.Name,
					Avatar:             userData.AvatarURL,
					IdentityProvider:   "GitHub",
					IdentityProviderID: fmt.Sprintf("%d", userData.ID),
				})
				if err != nil {
					logger.Error("error creating user", "email", email, "err", err.Error())
					http.Error(w, "something went wrong", http.StatusInternalServerError)
					return
				}
				logger.Info("user signed up", "email", email)
			} else {
				logger.Error("unexpected error getting user by email", "email", email, "err", err.Error())
				http.Error(w, "something went wrong", http.StatusInternalServerError)
				return
			}
		} else {
			logger.Info("user signed in", "email", email)
		}

		jwt, err := authService.Generate(user.ID.String(), userData.AvatarURL)
		if err != nil {
			logger.Error("error generating JWT", "err", err.Error())
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     AuthCookieName,
			Value:    jwt,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		}

		err = cookies.Write(w, cookie)
		if err != nil {
			logger.Error("error writing cookie", "err", err.Error())
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
	}
}
