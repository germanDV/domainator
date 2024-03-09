package middleware

import "net/http"

func Common(next http.Handler) http.Handler {
	return RealIP(Logger(RateLimiter(Recover(Helmet(CSP(ContentType(next)))))))
}
