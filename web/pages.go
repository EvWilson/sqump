package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/handlers"

	"github.com/go-chi/chi/v5"
)

func (r *Router) showHome(w http.ResponseWriter, req *http.Request) {
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
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
		Name        string
		Requests    []data.Request
	}
	info := make([]fileInfo, 0, len(files))
	for _, path := range files {
		sq, err := data.ReadCollection(path)
		if err != nil {
			r.ServerError(w, err)
			return
		}
		info = append(info, fileInfo{
			EscapedPath: url.PathEscape(strings.TrimPrefix(path, "/")),
			Name:        sq.Name,
			Requests:    sq.Requests,
		})
	}
	r.Render(w, 200, "home.tmpl.html", struct {
		CoreEnvironmentText string
		CurrentEnvironment  string
		Files               []fileInfo
		Error               string
	}{
		CoreEnvironmentText: string(envBytes),
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
	sq, err := data.ReadCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	envBytes, err := json.MarshalIndent(sq.Environment, "", "  ")
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "collection.tmpl.html", struct {
		Name               string
		EscapedPath        string
		EnvironmentText    string
		CurrentEnvironment string
		Requests           []data.Request
		Error              string
	}{
		Name:               sq.Name,
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
	name, ok := getParamEscaped(r, w, req, "name")
	if !ok {
		return
	}
	sq, err := data.ReadCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf, err := data.ReadConfigFrom(data.DefaultConfigLocation())
	if err != nil {
		r.ServerError(w, err)
		return
	}
	scope := req.URL.Query().Get("scope")
	var envMap data.EnvMap
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
	request, ok := sq.GetRequest(name)
	if !ok {
		r.ServerError(w, fmt.Errorf("no request '%s' found in collection '%s'", name, sq.Name))
		return
	}
	r.Render(w, 200, "request.tmpl.html", struct {
		EscapedPath        string
		CollectionName     string
		Name               string
		EditText           string
		EnvironmentText    string
		CurrentEnvironment string
		ExecText           string
		EnvScope           string
		Error              string
	}{
		EscapedPath:        url.PathEscape(path),
		CollectionName:     sq.Name,
		Name:               name,
		EditText:           request.Script.String(),
		EnvironmentText:    string(envBytes),
		CurrentEnvironment: conf.CurrentEnv,
		ExecText:           "",
		EnvScope:           scope,
		Error:              GetError(w, req),
	})
}

func (r *Router) showRenameCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	sq, err := data.ReadCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "renameCollection.tmpl.html", struct {
		EscapedPath string
		Name        string
		Error       string
	}{
		EscapedPath: url.PathEscape(path),
		Name:        sq.Name,
		Error:       GetError(w, req),
	})
}

func (r *Router) showUnregisterCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	sq, err := data.ReadCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "unregister.tmpl.html", struct {
		EscapedPath string
		Name        string
		Error       string
	}{
		EscapedPath: url.PathEscape(path),
		Name:        sq.Name,
		Error:       GetError(w, req),
	})
}

func (r *Router) showDeleteCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	sq, err := data.ReadCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "deleteCollection.tmpl.html", struct {
		EscapedPath string
		Name        string
		Error       string
	}{
		EscapedPath: url.PathEscape(path),
		Name:        sq.Name,
		Error:       GetError(w, req),
	})
}

func (r *Router) showRenameRequest(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	name, ok := getParamEscaped(r, w, req, "name")
	if !ok {
		return
	}
	sq, err := data.ReadCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "renameRequest.tmpl.html", struct {
		EscapedPath    string
		CollectionName string
		RequestName    string
		Error          string
	}{
		EscapedPath:    url.PathEscape(path),
		CollectionName: sq.Name,
		RequestName:    name,
		Error:          GetError(w, req),
	})
}

func (r *Router) showDeleteRequest(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	name, ok := getParamEscaped(r, w, req, "name")
	if !ok {
		return
	}
	sq, err := data.ReadCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	r.Render(w, 200, "deleteRequest.tmpl.html", struct {
		EscapedPath    string
		CollectionName string
		RequestName    string
		Previous       string
		Error          string
	}{
		EscapedPath:    url.PathEscape(path),
		CollectionName: sq.Name,
		RequestName:    name,
		Error:          GetError(w, req),
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
