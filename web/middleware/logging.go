package middleware

import (
	"net/http"

	"github.com/EvWilson/sqump/web/log"
)

func LoggingMiddleware(l log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Debug("incoming request", "remote", r.RemoteAddr, "method", r.Method, "path", r.URL.RequestURI())
			next.ServeHTTP(w, r)
		})
	}
}
