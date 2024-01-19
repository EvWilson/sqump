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

func (r *Router) showHome(w http.ResponseWriter, _ *http.Request) {
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
		Files               []fileInfo
	}{
		BaseEnvironmentText: string(envBytes),
		Files:               info,
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
	r.Render(w, 200, "collection.tmpl.html", struct {
		Title           string
		EscapedPath     string
		EnvironmentText string
		Requests        []core.Request
	}{
		Title:           sq.Title,
		EscapedPath:     url.PathEscape(path),
		EnvironmentText: string(envBytes),
		Requests:        sq.Requests,
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
	envBytes, err := json.MarshalIndent(sq.Environment, "", "  ")
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
		EscapedPath     string
		CollectionTitle string
		Title           string
		EditText        string
		EnvironmentText string
		ExecText        string
	}{
		EscapedPath:     url.PathEscape(path),
		CollectionTitle: sq.Title,
		Title:           title,
		EditText:        request.Script.String(),
		EnvironmentText: string(envBytes),
		ExecText:        "",
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
