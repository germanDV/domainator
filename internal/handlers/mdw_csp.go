package handlers

import (
	"fmt"
	"net/http"
)


func csp(next http.Handler) http.Handler {
	htmxCSSHash := "sha256-pgn1TCGZX6O77zDvy0oTODMOxemn0oj0LeCnQTRj7Kg="

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cspHeader := fmt.Sprintf("default-src 'self'; script-src 'self'; style-src 'self' '%s'; frame-ancestors 'none'; form-action 'self'", htmxCSSHash)

		w.Header().Set("Content-Security-Policy", cspHeader)
		next.ServeHTTP(w, r)
	})
}
