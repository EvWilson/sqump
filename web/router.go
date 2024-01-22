package web

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/EvWilson/sqump/web/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	*chi.Mux
	l         log.Logger
	logLevel  slog.Level
	templates TemplateCache
}

func NewRouter() (*Router, error) {
	mux := chi.NewMux()
	logLevel := slogLevellFromEnv()
	l := log.NewLogger(logLevel)
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
		logLevel:  logLevel,
		templates: tc,
	}
	mux.Get("/", r.showHome)
	mux.Post("/config", r.handleBaseConfig)
	mux.Get("/ws", r.handleSocketConnection)
	mux.Post("/autoregister", r.performAutoregister)
	mux.Post("/collection/create/new", r.createCollection)
	mux.Route("/collection/{path}", func(mux chi.Router) {
		mux.Get("/", r.showCollection)
		mux.Post("/config", r.handleCollectionConfig)
		mux.Get("/unregister", r.showUnregisterCollection)
		mux.Post("/unregister", r.performUnregisterCollection)
		mux.Route("/request", func(mux chi.Router) {
			mux.Post("/create/new", r.createRequest)
			mux.Get("/{title}", r.showRequest)
			mux.Post("/{title}/edit-script", r.updateRequestScript)
			mux.Get("/{title}/delete", r.showDeleteRequest)
			mux.Post("/{title}/delete", r.performDeleteRequest)
		})
	})
	subAssets, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}
	mux.Get("/*", http.FileServer(http.FS(subAssets)).ServeHTTP)
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
	if r.logLevel < slog.LevelInfo {
		r.l.Error(err.Error(), "stack", string(debug.Stack()))
	} else {
		r.l.Error(err.Error())
	}
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

func slogLevellFromEnv() slog.Level {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
