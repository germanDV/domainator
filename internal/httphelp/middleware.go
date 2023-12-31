package httphelp

import (
	"context"
	"domainator/internal/config"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/justinas/alice"
)

var (
	// Base is the base middleware that attaches auth info to the context
	Base = alice.New(authenticate)
	// Protected builds on Base and requires authentication
	Protected = Base.Append(requireAuth)
)

// Standard is the middleware that applies to all requests
func Standard(logger *slog.Logger) alice.Chain {
	logRequest := createLogger(logger)
	return alice.New(recoverPanic, logRequest, secureHeaders)
}

func createLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info(
				"received request",
				"ip", r.RemoteAddr,
				"proto", r.Proto,
				"method", r.Method,
				"url", r.URL.RequestURI(),
			)
			next.ServeHTTP(w, r)
		})
	}
}

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				ServerError(w, fmt.Errorf("recovered from panic. %v", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

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

func authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || cookie == nil {
			next.ServeHTTP(w, r)
			return
		}

		err = cookie.Valid()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		if cookie.Value == "" {
			next.ServeHTTP(w, r)
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.GetPublicKey(), nil
		})
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := uuid.Parse(claims["sub"].(string))
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)

		planID, ok := claims["pln"].(float64)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		ctx = context.WithValue(ctx, PlanIDContextKey, int(planID))

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := GetUserIDFromCtx(w, r)
		if userID == uuid.Nil {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		w.Header().Add("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
