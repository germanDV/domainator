package middleware

import "net/http"

func Common(next http.Handler) http.Handler {
	return Logger(ContentType(next))
}
