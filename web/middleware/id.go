package middleware

import (
	"net/http"

	"github.com/EvWilson/sqump/web/util"
)

func IDHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := util.GetID(r); err != nil {
			_ = util.SetNewID(w)
		}
		next.ServeHTTP(w, r)
	})
}
