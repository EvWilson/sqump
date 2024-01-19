package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/EvWilson/sqump/core"
	"github.com/EvWilson/sqump/handlers"
)

func (r *Router) handleBaseConfig(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	envMap, err := configMap(req)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf.Environment = envMap
	err = conf.Flush()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func (r *Router) handleCollectionConfig(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	titleSlice, ok := req.Form["title"]
	if !ok {
		r.RequestError(w, errors.New("save request config form does not contain field 'title'"))
		return
	}
	title := strings.Join(titleSlice, "\n")
	envMap, err := configMap(req)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	sq, err := core.ReadSqumpfile("/" + path)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	sq.Environment = envMap
	err = sq.Flush()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, fmt.Sprintf("/collection/%s/request/%s", url.PathEscape(path), title), http.StatusFound)
}

func (r *Router) createCollection(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	reqTitle, ok := req.Form["title"]
	if !ok {
		r.RequestError(w, errors.New("create collection form does not contain field 'title'"))
		return
	}
	title := strings.Join(reqTitle, "\n")
	err = handlers.AddFile(title)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func (r *Router) performUnregisterCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	err := handlers.Unregister("/" + path)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func (r *Router) createRequest(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	reqTitle, ok := req.Form["title"]
	if !ok {
		r.RequestError(w, errors.New("create request form does not contain field 'title'"))
		return
	}
	title := strings.Join(reqTitle, "\n")
	err = handlers.AddRequest("/"+path, title)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, fmt.Sprintf("/collection/%s/request/%s", url.PathEscape(path), title), http.StatusFound)
}

func (r *Router) updateRequestScript(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	title, ok := getParamEscaped(r, w, req, "title")
	if !ok {
		return
	}
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	script, ok := req.Form["edit"]
	if !ok {
		r.RequestError(w, errors.New("update request form does not contain field 'edit'"))
		return
	}
	err = handlers.UpdateRequestScript("/"+path, title, script)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, fmt.Sprintf("/collection/%s/request/%s", url.PathEscape(path), title), http.StatusFound)
}

func (r *Router) performDeleteRequest(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	title, ok := getParamEscaped(r, w, req, "title")
	if !ok {
		return
	}
	err := handlers.RemoveRequest("/"+path, title)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, fmt.Sprintf("/collection/%s", url.PathEscape(path)), http.StatusFound)
}

func (r *Router) performAutoregister(w http.ResponseWriter, req *http.Request) {
	cwd, err := os.Getwd()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	err = handlers.Autoregister(cwd)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func configMap(req *http.Request) (core.EnvMap, error) {
	configData, ok := req.Form["config"]
	if !ok {
		return nil, errors.New("base config form does not contain field 'config'")
	}
	configString := strings.Join(configData, "\n")
	var envMap core.EnvMap
	err := json.Unmarshal([]byte(configString), &envMap)
	if err != nil {
		return nil, err
	}
	return envMap, nil
}
