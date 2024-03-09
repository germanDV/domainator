package middleware

import "net/http"

func Common(next http.Handler) http.Handler {
	return Logger(Recover(CSP(ContentType(next))))
}
