package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/EvWilson/sqump/core"
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
	http.Redirect(w, req, req.Header.Get("Referer"), http.StatusFound)
}

func (r *Router) handleCollectionConfig(w http.ResponseWriter, req *http.Request) {
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
	http.Redirect(w, req, req.Header.Get("Referer"), http.StatusFound)
}

func configMap(req *http.Request) (core.EnvMap, error) {
	configData, ok := req.Form["config"]
	if !ok {
		return nil, fmt.Errorf("base config form does not contain field 'config'")
	}
	configString := strings.Join(configData, "\n")
	var envMap core.EnvMap
	err := json.Unmarshal([]byte(configString), &envMap)
	if err != nil {
		return nil, err
	}
	return envMap, nil
}
