package handlers

import (
	"errors"
	"fmt"
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

func GithubLogin(githubConfig *oauth2.Config, cookieSigningSecret []byte) http.HandlerFunc {
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
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		url := githubConfig.AuthCodeURL(state)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func GithubCallback(
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
			http.Error(w, "Invalid or missing state", http.StatusUnauthorized)
			return
		}

		token, err := githubConfig.Exchange(r.Context(), code)
		if err != nil {
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		jwt, err := authService.Generate(user.ID.String(), userData.AvatarURL)
		if err != nil {
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
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
	}
}
