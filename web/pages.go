package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/EvWilson/sqump/data"
	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/prnt"
	"github.com/EvWilson/sqump/web/stores"
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
	info := make([]data.Collection, 0, len(files))
	for _, path := range files {
		coll, err := handlers.GetCollection(path)
		if err != nil {
			if errors.Is(err, data.ErrNotFound{}) {
				prnt.Println("WARNING: while loading home:", err.Error())
				continue
			}
			r.ServerError(w, err)
			return
		}
		info = append(info, *coll)
	}
	r.Render(w, 200, "home.tmpl.html", struct {
		CoreEnvironmentText string
		CurrentEnvironment  string
		Files               []data.Collection
		Error               string
	}{
		CurrentEnvironment: conf.CurrentEnv,
		Files:              info,
		Error:              util.GetErrorOnRequest(w, req),
	})
}

func (r *Router) showCollection(ces stores.CurrentEnvService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
		currentEnv, err := ces.GetCurrentEnv(req)
		if err != nil {
			r.ServerError(w, err)
			return
		}
		r.Render(w, 200, "collection.tmpl.html", struct {
			Name               string
			Path               string
			EnvironmentText    string
			CurrentEnvironment string
			Requests           []data.Request
			Error              string
		}{
			Name:               coll.Name,
			Path:               path,
			EnvironmentText:    string(envBytes),
			CurrentEnvironment: currentEnv,
			Requests:           coll.Requests,
			Error:              util.GetErrorOnRequest(w, req),
		})
	}
}

func (r *Router) showRequest(ces stores.CurrentEnvService, tcs stores.TempConfigService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
		scope := req.URL.Query().Get("scope")
		var envMap data.EnvMap
		switch scope {
		case "":
			fallthrough
		case "collection":
			envMap = coll.Environment
			scope = "Collection"
		case "temp":
			envMap, err = tcs.GetTempEnv(req)
			if err != nil {
				r.ServerError(w, err)
				return
			}
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
		currentEnv, err := ces.GetCurrentEnv(req)
		if err != nil {
			r.ServerError(w, err)
			return
		}
		r.Render(w, 200, "request.tmpl.html", struct {
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
			CollectionName:     coll.Name,
			CollectionPath:     coll.Path,
			Requests:           coll.Requests,
			Name:               name,
			EditText:           request.Script.String(),
			EnvironmentText:    string(envBytes),
			CurrentEnvironment: currentEnv,
			ExecText:           "",
			EnvScope:           scope,
			Error:              util.GetErrorOnRequest(w, req),
		})
	}
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
