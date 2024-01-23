package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers"

	"github.com/go-chi/chi/v5"
)

func (r *Router) showHome(w http.ResponseWriter, req *http.Request) {
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		r.ServerError(w, err)
		return
	}
	envBytes, err := json.MarshalIndent(conf.Environment, "", "  ")
	if err != nil {
		r.ServerError(w, err)
		return
	}
	files, err := handlers.ListCollections()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	type fileInfo struct {
		EscapedPath string
		Title       string
		Requests    []core.Request
	}
	info := make([]fileInfo, 0, len(files))
	for _, path := range files {
		sq, err := core.ReadSqumpfile(path)
		if err != nil {
			r.ServerError(w, err)
			return
		}
		info = append(info, fileInfo{
			EscapedPath: url.PathEscape(strings.TrimPrefix(path, "/")),
			Title:       sq.Title,
			Requests:    sq.Requests,
		})
	}
	r.Render(w, 200, "home.tmpl.html", struct {
		BaseEnvironmentText string
		CurrentEnvironment  string
		Files               []fileInfo
		Error               string
	}{
		BaseEnvironmentText: string(envBytes),
		CurrentEnvironment:  conf.CurrentEnv,
		Files:               info,
		Error:               GetError(w, req),
	})
}

func (r *Router) showCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	sq, err := core.ReadSqumpfile(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	envBytes, err := json.MarshalIndent(sq.Environment, "", "  ")
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "collection.tmpl.html", struct {
		Title              string
		EscapedPath        string
		EnvironmentText    string
		CurrentEnvironment string
		Requests           []core.Request
		Error              string
	}{
		Title:              sq.Title,
		EscapedPath:        url.PathEscape(path),
		EnvironmentText:    string(envBytes),
		CurrentEnvironment: conf.CurrentEnv,
		Requests:           sq.Requests,
		Error:              GetError(w, req),
	})
}

func (r *Router) showRequest(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	title, ok := getParamEscaped(r, w, req, "title")
	if !ok {
		return
	}
	sq, err := core.ReadSqumpfile("/" + path)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		r.ServerError(w, err)
		return
	}
	scope := req.URL.Query().Get("scope")
	var envMap core.EnvMap
	switch scope {
	case "":
		fallthrough
	case "collection":
		envMap = sq.Environment
		scope = "Collection"
	case "core":
		envMap = conf.Environment
		scope = "Core"
	case "temp":
		envMap = getTempConfig()
		scope = "Temporary"
	default:
		r.RequestError(w, fmt.Errorf("unrecognized scope '%s'", scope))
		return
	}
	envBytes, err := json.MarshalIndent(envMap, "", "  ")
	if err != nil {
		r.ServerError(w, err)
		return
	}
	request, ok := sq.GetRequest(title)
	if !ok {
		r.ServerError(w, fmt.Errorf("no request '%s' found in squmpfile '%s'", title, sq.Title))
		return
	}
	r.Render(w, 200, "request.tmpl.html", struct {
		EscapedPath        string
		CollectionTitle    string
		Title              string
		EditText           string
		EnvironmentText    string
		CurrentEnvironment string
		ExecText           string
		EnvScope           string
		Error              string
	}{
		EscapedPath:        url.PathEscape(path),
		CollectionTitle:    sq.Title,
		Title:              title,
		EditText:           request.Script.String(),
		EnvironmentText:    string(envBytes),
		CurrentEnvironment: conf.CurrentEnv,
		ExecText:           "",
		EnvScope:           scope,
		Error:              GetError(w, req),
	})
}

func (r *Router) showUnregisterCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	sq, err := core.ReadSqumpfile("/" + path)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "unregister.tmpl.html", struct {
		EscapedPath string
		Title       string
	}{
		EscapedPath: url.PathEscape(path),
		Title:       sq.Title,
	})
}

func (r *Router) showDeleteCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	sq, err := core.ReadSqumpfile("/" + path)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "deleteCollection.tmpl.html", struct {
		EscapedPath string
		Title       string
	}{
		EscapedPath: url.PathEscape(path),
		Title:       sq.Title,
	})
}

func (r *Router) showDeleteRequest(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	title, ok := getParamEscaped(r, w, req, "title")
	if !ok {
		return
	}
	sq, err := core.ReadSqumpfile("/" + path)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "deleteRequest.tmpl.html", struct {
		EscapedPath     string
		CollectionTitle string
		RequestTitle    string
		Previous        string
	}{
		EscapedPath:     url.PathEscape(path),
		CollectionTitle: sq.Title,
		RequestTitle:    title,
	})
}

func getParamEscaped(r *Router, w http.ResponseWriter, req *http.Request, key string) (string, bool) {
	param := chi.URLParam(req, key)
	if param == "" {
		r.RequestError(w, fmt.Errorf("no '%s' param found in '%s'", key, req.URL.Path))
		return "", false
	}
	param, err := url.PathUnescape(param)
	if err != nil {
		r.ServerError(w, err)
		return "", false
	}
	return param, true
}
