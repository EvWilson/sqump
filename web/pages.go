package web

import "net/http"

func (r *Router) home() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.Render(w, http.StatusOK, "index.tmpl.html", nil)
	})
}
