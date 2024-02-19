package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"slices"
)

var allowedStatuses = []int{http.StatusOK, http.StatusFound}

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &ErrorResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		if !isAllowedStatus(rw.statusCode) {
			referer := r.Header.Get("Referer")
			if referer == "" {
				referer = "/"
			}
			http.Redirect(w, r, referer, http.StatusFound)
		}
	})
}

type ErrorResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *ErrorResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	if isAllowedStatus(code) {
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *ErrorResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return h.Hijack()
}

func isAllowedStatus(status int) bool {
	return slices.Contains(allowedStatuses, status)
}
