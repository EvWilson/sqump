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

	"github.com/EvWilson/sqump/prnt"
	"github.com/EvWilson/sqump/web/log"
	"github.com/EvWilson/sqump/web/middleware"
	"github.com/EvWilson/sqump/web/util"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	*chi.Mux
	l         log.Logger
	logLevel  slog.Level
	templates TemplateCache
}

func NewRouter(isReadonly bool) (*Router, error) {
	if isReadonly {
		prnt.Println("starting server in readonly mode")
	}
	mux := chi.NewMux()
	logLevel := slogLevellFromEnv()
	l := log.NewLogger(logLevel)
	mux.Use(
		chiMiddleware.Recoverer,
		middleware.ErrorHandler,
		middleware.ReadonlyMiddleware(isReadonly),
		middleware.LoggingMiddleware(l),
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
	mux.Post("/current-env", r.setCurrentEnv)
	mux.Get("/ws", r.handleSocketConnection())
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
			mux.Get("/{name}", r.showRequest)
			mux.Post("/{name}/edit-script", r.updateRequestScript)
			mux.Get("/{name}/rename", r.showRenameRequest)
			mux.Post("/{name}/rename", r.handleRenameRequest)
			mux.Get("/{name}/delete", r.showDeleteRequest)
			mux.Post("/{name}/delete", r.performDeleteRequest)
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
	util.SetErrorCookie(w, fmt.Sprintf("Server error: %v", err))
	Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (r *Router) RequestError(w http.ResponseWriter, err error) {
	r.l.Error(err.Error(), "stack", string(debug.Stack()))
	util.SetErrorCookie(w, fmt.Sprintf("Request error: %v", err))
	Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
}

func Error(w http.ResponseWriter, error string, code int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
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
