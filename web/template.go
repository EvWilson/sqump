package web

import (
	"fmt"
	"net/url"
	"text/template"
)

type TemplateCache map[string]*template.Template

func NewTemplateCache() (TemplateCache, error) {
	cache := TemplateCache{}
	pages, err := assets.ReadDir("assets/templates/pages")
	if err != nil {
		return nil, err
	}
	templFuncs := template.FuncMap{
		"pathescape": url.PathEscape,
	}
	for _, page := range pages {
		files := []string{
			"assets/templates/index.tmpl.html",
			fmt.Sprintf("assets/templates/pages/%s", page.Name()),
		}
		ts, err := template.New(page.Name()).Funcs(templFuncs).ParseFS(assets, files...)
		if err != nil {
			return nil, err
		}
		cache[page.Name()] = ts
	}
	return cache, nil
}
