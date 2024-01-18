package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/EvWilson/sqump/web/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	*chi.Mux
	l         log.Logger
	templates TemplateCache
}

func NewRouter() (*Router, error) {
	mux := chi.NewMux()
	l := log.NewLogger(slog.LevelInfo)
	mux.Use(
		middleware.Recoverer,
		LoggingMiddleware(l),
	)
	tc, err := NewTemplateCache()
	if err != nil {
		return nil, err
	}
	r := &Router{
		Mux:       mux,
		l:         l,
		templates: tc,
	}
	mux.Get("/", r.home)
	mux.Post("/config", r.handleBaseConfig)
	mux.Get("/ws", r.handleSocketConnection)
	mux.Route("/collection/{path}", func(mux chi.Router) {
		mux.Get("/", r.collection)
		mux.Post("/config", r.handleCollectionConfig)
		mux.Get("/request/{title}", r.request)
	})
	mux.Get("/*", http.FileServer(http.Dir("./web/assets")).ServeHTTP)
	return r, nil
}

func (r *Router) Render(w http.ResponseWriter, status int, page string, data any) {
	ts, ok := r.templates[page]
	if !ok {
		r.ServerError(w, fmt.Errorf("could not find template of name: %s", page))
		return
	}
	w.WriteHeader(status)
	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		r.ServerError(w, err)
	}
}

func (r *Router) ServerError(w http.ResponseWriter, err error) {
	r.l.Error(err.Error(), "stack", string(debug.Stack()))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (r *Router) RequestError(w http.ResponseWriter, err error) {
	r.l.Error(err.Error(), "stack", string(debug.Stack()))
	http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func LoggingMiddleware(l log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Debug("incoming request", "remote", r.RemoteAddr, "method", r.Method, "path", r.URL.RequestURI())
			next.ServeHTTP(w, r)
		})
	}
}
