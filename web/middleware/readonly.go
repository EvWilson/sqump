package middleware

import (
	"net/http"

	"github.com/EvWilson/sqump/web/util"
)

func ReadonlyMiddleware(isReadonly bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isReadonly && r.Method != "GET" {
				w.WriteHeader(http.StatusForbidden)
				util.SetErrorCookie(w, "Request Error: method not allowed on this route in readonly mode")
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
