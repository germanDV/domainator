package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

type contextKey string

var (
	HtmxNonceKey     = contextKey("htmxNonce")
	RespTrgtNonceKey = contextKey("respTrgtNonce")
	StylesNonceKey   = contextKey("stylesNonce")
)

func generateRandomString(len int) string {
	bytes := make([]byte, len)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func CSP(next http.Handler) http.Handler {
	// The hash of the CSS that HTMX injects
	htmxCSSHash := "sha256-pgn1TCGZX6O77zDvy0oTODMOxemn0oj0LeCnQTRj7Kg="

	htmxNonce := generateRandomString(32)
	respTrgtNonce := generateRandomString(32)
	stylesNonce := generateRandomString(32)

	ctx := context.WithValue(context.Background(), HtmxNonceKey, htmxNonce)
	ctx = context.WithValue(ctx, RespTrgtNonceKey, respTrgtNonce)
	ctx = context.WithValue(ctx, StylesNonceKey, stylesNonce)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cspHeader := fmt.Sprintf(
			"default-src 'self'; script-src 'nonce-%s' 'nonce-%s' 'unsafe-eval'; style-src 'nonce-%s' '%s';",
			htmxNonce, respTrgtNonce, stylesNonce, htmxCSSHash,
		)

		w.Header().Set("Content-Security-Policy", cspHeader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
