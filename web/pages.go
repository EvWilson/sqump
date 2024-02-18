package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/web/util"

	"github.com/go-chi/chi/v5"
)

func (r *Router) showHome(w http.ResponseWriter, req *http.Request) {
	conf, err := handlers.GetConfig()
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
		coll, err := handlers.GetCollection(path)
		if err != nil {
			r.ServerError(w, err)
			return
		}
		info = append(info, fileInfo{
			EscapedPath: url.PathEscape(strings.TrimPrefix(path, "/")),
			Name:        coll.Name,
			Requests:    coll.Requests,
		})
	}
	r.Render(w, 200, "home.tmpl.html", struct {
		CoreEnvironmentText string
		CurrentEnvironment  string
		Files               []fileInfo
		Error               string
	}{
		CurrentEnvironment: conf.CurrentEnv,
		Files:              info,
		Error:              util.GetErrorOnRequest(w, req),
	})
}

func (r *Router) showCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	coll, err := handlers.GetCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	envBytes, err := json.MarshalIndent(coll.Environment, "", "  ")
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf, err := handlers.GetConfig()
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
		Name:               coll.Name,
		EscapedPath:        url.PathEscape(path),
		EnvironmentText:    string(envBytes),
		CurrentEnvironment: conf.CurrentEnv,
		Requests:           coll.Requests,
		Error:              util.GetErrorOnRequest(w, req),
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
	coll, err := handlers.GetCollection(fmt.Sprintf("/%s", path))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf, err := handlers.GetConfig()
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
		envMap = coll.Environment
		scope = "Collection"
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
	request, ok := coll.GetRequest(name)
	if !ok {
		r.ServerError(w, fmt.Errorf("no request '%s' found in collection '%s'", name, coll.Name))
		return
	}
	r.Render(w, 200, "request.tmpl.html", struct {
		EscapedPath        string
		CollectionName     string
		CollectionPath     string
		Requests           []data.Request
		Name               string
		EditText           string
		EnvironmentText    string
		CurrentEnvironment string
		ExecText           string
		EnvScope           string
		Error              string
	}{
		EscapedPath:        url.PathEscape(path),
		CollectionName:     coll.Name,
		CollectionPath:     url.PathEscape(coll.Path),
		Requests:           coll.Requests,
		Name:               name,
		EditText:           request.Script.String(),
		EnvironmentText:    string(envBytes),
		CurrentEnvironment: conf.CurrentEnv,
		ExecText:           "",
		EnvScope:           scope,
		Error:              util.GetErrorOnRequest(w, req),
	})
}

func (r *Router) showRenameCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	coll, err := handlers.GetCollection(fmt.Sprintf("/%s", path))
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
		Name:        coll.Name,
		Error:       util.GetErrorOnRequest(w, req),
	})
}

func (r *Router) showUnregisterCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	coll, err := handlers.GetCollection(fmt.Sprintf("/%s", path))
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
		Name:        coll.Name,
		Error:       util.GetErrorOnRequest(w, req),
	})
}

func (r *Router) showDeleteCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	coll, err := handlers.GetCollection(fmt.Sprintf("/%s", path))
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
		Name:        coll.Name,
		Error:       util.GetErrorOnRequest(w, req),
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
	coll, err := handlers.GetCollection(fmt.Sprintf("/%s", path))
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
		CollectionName: coll.Name,
		RequestName:    name,
		Error:          util.GetErrorOnRequest(w, req),
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
	coll, err := handlers.GetCollection(fmt.Sprintf("/%s", path))
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
		CollectionName: coll.Name,
		RequestName:    name,
		Error:          util.GetErrorOnRequest(w, req),
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
