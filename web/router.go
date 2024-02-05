package web

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/EvWilson/sqump/web/log"
	"github.com/EvWilson/sqump/web/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
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
		chiMiddleware.Recoverer,
		middleware.LoggingMiddleware(l),
		middleware.ErrorHandler,
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
	mux.Post("/current-env", r.setCurrentEnv)
	mux.Get("/ws", r.handleSocketConnection)
	mux.Post("/autoregister", r.performAutoregister)
	mux.Post("/collection/create/new", r.createCollection)
	mux.Route("/collection/{path}", func(mux chi.Router) {
		mux.Get("/", r.showCollection)
		mux.Post("/config", r.handleCollectionConfig)
		mux.Get("/rename", r.showRenameCollection)
		mux.Post("/rename", r.handleRenameCollection)
		mux.Get("/unregister", r.showUnregisterCollection)
		mux.Post("/unregister", r.handleUnregisterCollection)
		mux.Get("/delete", r.showDeleteCollection)
		mux.Post("/delete", r.handleDeleteCollection)
		mux.Route("/request", func(mux chi.Router) {
			mux.Post("/create/new", r.createRequest)
			mux.Get("/{title}", r.showRequest)
			mux.Post("/{title}/edit-script", r.updateRequestScript)
			mux.Get("/{title}/rename", r.showRenameRequest)
			mux.Post("/{title}/rename", r.handleRenameRequest)
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
	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		r.ServerError(w, err)
	}
}

func (r *Router) ServerError(w http.ResponseWriter, err error) {
	var pcs [1]uintptr
	_ = runtime.Callers(2, pcs[:])
	rec := slog.NewRecord(time.Now(), slog.LevelError, err.Error(), pcs[0])
	if r.logLevel < slog.LevelInfo {
		_ = r.l.With("error", err).Handler().Handle(context.Background(), rec)
	} else {
		_ = r.l.With("error", err).Handler().Handle(context.Background(), rec)
	}
	setErrorCookie(w, fmt.Sprintf("Server error: %v", err))
	Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (r *Router) RequestError(w http.ResponseWriter, err error) {
	r.l.Error(err.Error(), "stack", string(debug.Stack()))
	setErrorCookie(w, fmt.Sprintf("Request error: %v", err))
	Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func setErrorCookie(w http.ResponseWriter, err string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "error",
		Value:    err,
		Path:     "/",
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
}

func Error(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
}

func GetError(w http.ResponseWriter, req *http.Request) string {
	cookie, err := req.Cookie("error")
	if err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "error",
			Value:    "",
			MaxAge:   -1,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
		})
		return cookie.Value
	} else {
		return ""
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
