package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/githubauth"
	"github.com/germandv/domainator/internal/tokenauth"
	"github.com/germandv/domainator/internal/users"
	"golang.org/x/oauth2"
)

func GithubLogin(stateStr string, githubConfig *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)
		if userID != "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		url := githubConfig.AuthCodeURL(stateStr)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

// TODO: Save user profile img from GitHub and use it in layout.templ
func GithubCallback(
	stateStr string,
	githubConfig *oauth2.Config,
	authService tokenauth.Service,
	usersService users.Service,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := r.FormValue("state")
		code := r.FormValue("code")

		if state != stateStr {
			http.Error(w, "Invalid state", http.StatusUnauthorized)
			return
		}

		token, err := githubConfig.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		jwt, err := authService.Generate(user.ID.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:     AuthCookieName,
			Value:    jwt,
			Path:     "/",
			HttpOnly: true,
		}

		http.SetCookie(w, &cookie)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
