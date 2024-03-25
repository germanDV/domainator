package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/tokenauth"
)

const AuthCookieName = "token"

// AuthMdwBuilder returns a middleware that checks if there is an auth cookie,
// and if so decodes it and adds the user ID to the request context.
//
// If `required` is `true`, it will return an error if there is no auth cookie.
// If it is `false`, it will call the next handler.
func AuthMdwBuilder(auth tokenauth.Service, required bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(AuthCookieName)

			var e error
			var userID string

			if err == nil && cookie.Value != "" {
				claims, err := auth.Validate(cookie.Value)
				if err != nil {
					e = err
				} else {
					userID = claims["sub"].(string)
				}
			}

			if e == nil && userID != "" {
				r = cntxt.SetUserID(r, userID)
				next.ServeHTTP(w, r)
			} else {
				if required {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
				} else {
					next.ServeHTTP(w, r)
				}
			}
		})
	}
}
