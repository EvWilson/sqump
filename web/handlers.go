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
	err = saveCoreConfig(req)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func (r *Router) setCurrentEnv(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	envSlice, ok := req.Form["current"]
	if !ok {
		r.RequestError(w, errors.New("save current env form does not contain field 'current'"))
		return
	}
	env := strings.Join(envSlice, "\n")
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		r.ServerError(w, err)
		return
	}
	conf.CurrentEnv = env
	err = conf.Flush()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, req.Header.Get("Referer"), http.StatusFound)
}

func (r *Router) handleCollectionConfig(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	scopeSlice, ok := req.Form["scope"]
	if !ok {
		r.RequestError(w, errors.New("save request config form does not contain field 'scope'"))
		return
	}
	scope := strings.Join(scopeSlice, "\n")
	titleSlice, ok := req.Form["title"]
	if !ok {
		titleSlice = []string{}
	}
	title := strings.Join(titleSlice, "\n")
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	var redirectURL string
	if title != "" {
		redirectURL = fmt.Sprintf("/collection/%s/request/%s", url.PathEscape(path), title)
	} else {
		redirectURL = fmt.Sprintf("/collection/%s", url.PathEscape(path))
	}
	switch scope {
	case "core":
		err = saveCoreConfig(req)
		if err != nil {
			r.ServerError(w, err)
			return
		}
		http.Redirect(w, req, fmt.Sprintf("%s?scope=core", redirectURL), http.StatusFound)
		return
	case "collection":
		envMap, err := configMap(req)
		if err != nil {
			r.ServerError(w, err)
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
		http.Redirect(w, req, redirectURL, http.StatusFound)
		return
	case "temp":
		err := saveTempConfig(req)
		if err != nil {
			r.ServerError(w, err)
			return
		}
		http.Redirect(w, req, fmt.Sprintf("%s?scope=temp", redirectURL), http.StatusFound)
		return
	default:
		r.RequestError(w, fmt.Errorf("unrecognized collection scope '%s'", scope))
		return
	}
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

func (r *Router) handleRenameCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	newTitle, ok := req.Form["new-name"]
	if !ok {
		r.RequestError(w, errors.New("rename collection form does not contain field 'new-name'"))
		return
	}
	err = handlers.UpdateCollectionName(fmt.Sprintf("/%s", path), strings.Join(newTitle, "\n"))
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func (r *Router) handleUnregisterCollection(w http.ResponseWriter, req *http.Request) {
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

func (r *Router) handleDeleteCollection(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	err := handlers.RemoveCollection("/" + path)
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

func (r *Router) handleRenameRequest(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	oldNameSlice, ok := req.Form["old-name"]
	if !ok {
		r.RequestError(w, errors.New("rename request form does not contain field 'old-name'"))
		return
	}
	oldName := strings.Join(oldNameSlice, "\n")
	newNameSlice, ok := req.Form["new-name"]
	if !ok {
		r.RequestError(w, errors.New("rename request form does not contain field 'new-name'"))
		return
	}
	newName := strings.Join(newNameSlice, "\n")
	err = handlers.UpdateRequestName(fmt.Sprintf("/%s", path), oldName, newName)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, fmt.Sprintf("/collection/%s/request/%s", url.PathEscape(path), newName), http.StatusFound)
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

func saveCoreConfig(req *http.Request) error {
	envMap, err := configMap(req)
	if err != nil {
		return err
	}
	conf, err := core.ReadConfigFrom(core.DefaultConfigLocation())
	if err != nil {
		return err
	}
	conf.Environment = envMap
	err = conf.Flush()
	if err != nil {
		return err
	}
	return nil
}
