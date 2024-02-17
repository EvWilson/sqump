package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &ErrorResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		if rw.statusCode != http.StatusOK {
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
	if code == http.StatusOK {
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
