package web

import (
	"path/filepath"
	"text/template"
)

type TemplateCache map[string]*template.Template

func NewTemplateCache() (TemplateCache, error) {
	cache := TemplateCache{}

	pages, err := filepath.Glob("./assets/html/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		files := []string{
			"./assetsindex.tmpl.html",
			page,
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}
