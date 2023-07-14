package main

import (
	"context"
	"domainator/internal/config"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; font-src 'self'")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		app.logit.Info(msg)
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || cookie == nil {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		err = cookie.Valid()
		if err != nil {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		if cookie.Value == "" {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.GetString("JWT_SECRET")), nil
		})
		if err != nil {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}

		w.Header().Add("Cache-Control", "no-store")
		ctx := context.WithValue(r.Context(), userIDContextKey, claims["sub"])
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
