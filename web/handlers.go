package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/EvWilson/sqump/handlers"
	"github.com/EvWilson/sqump/web/middleware"
	"github.com/EvWilson/sqump/web/stores"
	"github.com/EvWilson/sqump/web/util"
)

func (r *Router) setCurrentEnv(ces stores.CurrentEnvService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
		err = ces.SetCurrentEnv(req, env)
		if err != nil {
			r.ServerError(w, err)
			return
		}
		http.Redirect(w, req, req.Header.Get("Referer"), http.StatusFound)
	}
}

func (r *Router) handleCollectionConfig(isReadonly bool, tcs stores.TempConfigService) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
		nameSlice, ok := req.Form["name"]
		if !ok {
			nameSlice = []string{}
		}
		name := strings.Join(nameSlice, "\n")
		path, ok := getParamEscaped(r, w, req, "path")
		if !ok {
			return
		}
		var redirectURL string
		if name != "" {
			redirectURL = fmt.Sprintf("/collection/%s/request/%s", url.PathEscape(path), name)
		} else {
			redirectURL = fmt.Sprintf("/collection/%s", url.PathEscape(path))
		}
		switch scope {
		case "collection":
			if isReadonly {
				middleware.HandleReadonlyCondition(w)
				return
			}
			envMap, err := util.ConfigMap(req)
			if err != nil {
				r.ServerError(w, err)
				return
			}
			err = handlers.UpdateCollectionEnv(fmt.Sprintf("/%s", path), envMap)
			if err != nil {
				r.ServerError(w, err)
				return
			}
			http.Redirect(w, req, redirectURL, http.StatusFound)
			return
		case "override":
			err := tcs.SaveTempConfig(req)
			if err != nil {
				r.ServerError(w, err)
				return
			}
			http.Redirect(w, req, fmt.Sprintf("%s?scope=override", redirectURL), http.StatusFound)
			return
		default:
			r.RequestError(w, fmt.Errorf("unrecognized collection scope '%s'", scope))
			return
		}
	}
}

func (r *Router) createCollection(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		r.ServerError(w, err)
		return
	}
	reqName, ok := req.Form["name"]
	if !ok {
		r.RequestError(w, errors.New("create collection form does not contain field 'name'"))
		return
	}
	name := strings.Join(reqName, "\n")
	err = handlers.AddFile(name)
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
	newName, ok := req.Form["new-name"]
	if !ok {
		r.RequestError(w, errors.New("rename collection form does not contain field 'new-name'"))
		return
	}
	err = handlers.UpdateCollectionName(fmt.Sprintf("/%s", path), strings.Join(newName, "\n"))
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
	err := handlers.Unregister(fmt.Sprintf("/%s", path))
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
	err := handlers.RemoveCollection(fmt.Sprintf("/%s", path))
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
	reqName, ok := req.Form["name"]
	if !ok {
		r.RequestError(w, errors.New("create request form does not contain field 'name'"))
		return
	}
	name := strings.Join(reqName, "\n")
	err = handlers.AddRequest(fmt.Sprintf("/%s", path), name)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, fmt.Sprintf("/collection/%s/request/%s", url.PathEscape(path), name), http.StatusFound)
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
	name, ok := getParamEscaped(r, w, req, "name")
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
	err = handlers.UpdateRequestScript(fmt.Sprintf("/%s", path), name, script)
	if err != nil {
		r.ServerError(w, err)
		return
	}
	http.Redirect(w, req, fmt.Sprintf("/collection/%s/request/%s", url.PathEscape(path), name), http.StatusFound)
}

func (r *Router) performDeleteRequest(w http.ResponseWriter, req *http.Request) {
	path, ok := getParamEscaped(r, w, req, "path")
	if !ok {
		return
	}
	name, ok := getParamEscaped(r, w, req, "name")
	if !ok {
		return
	}
	err := handlers.RemoveRequest(fmt.Sprintf("/%s", path), name)
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
