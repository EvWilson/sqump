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
	"github.com/EvWilson/sqump/web/stores"
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
		middleware.IDHandler,
		middleware.ErrorHandler,
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

	subAssets, err := fs.Sub(assets, "assets")
	if err != nil {
		return nil, err
	}

	ces := stores.NewCurrentEnvService(isReadonly)
	tcs := stores.NewTempConfigService(ces)
	eps := stores.NewExecProxyService(ces, tcs)

	// These routes have a special case with the readonly mode
	mux.Group(func(plainMux chi.Router) {
		plainMux.Post("/current-env", r.setCurrentEnv(ces))
		plainMux.Post("/collection/{path}/config", r.handleCollectionConfig(isReadonly, tcs))
	})

	// These obey normal readonly mode rules
	mux.Group(func(roMux chi.Router) {
		roMux.Use(middleware.ReadonlyMiddleware(isReadonly))
		roMux.Get("/", r.showHome)
		roMux.Get("/ws", r.handleSocketConnection(eps))
		roMux.Post("/autoregister", r.performAutoregister)
		roMux.Post("/collection/create/new", r.createCollection)
		roMux.Route("/collection/{path}", func(roMux chi.Router) {
			roMux.Get("/", r.showCollection(ces))
			roMux.Get("/rename", r.showRenameCollection)
			roMux.Post("/rename", r.handleRenameCollection)
			roMux.Get("/unregister", r.showUnregisterCollection)
			roMux.Post("/unregister", r.handleUnregisterCollection)
			roMux.Get("/delete", r.showDeleteCollection)
			roMux.Post("/delete", r.handleDeleteCollection)
			roMux.Route("/request", func(roMux chi.Router) {
				roMux.Post("/create/new", r.createRequest)
				roMux.Get("/{name}", r.showRequest(ces, tcs))
				roMux.Post("/{name}/edit-script", r.updateRequestScript)
				roMux.Get("/{name}/rename", r.showRenameRequest)
				roMux.Post("/{name}/rename", r.handleRenameRequest)
				roMux.Get("/{name}/delete", r.showDeleteRequest)
				roMux.Post("/{name}/delete", r.performDeleteRequest)
			})
		})
		roMux.Get("/*", http.FileServer(http.FS(subAssets)).ServeHTTP)
	})

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
