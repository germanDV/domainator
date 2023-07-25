package main

import (
	"html/template"
	"path/filepath"
	"time"
)

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.ParseFiles("./ui/html/base.html.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.html.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/fragments/*.html.tmpl")
		if err != nil {
			return nil, err
		}

		ts, err = ts.Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func newFragmentsCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	fragments, err := filepath.Glob("./ui/html/fragments/*.html.tmpl")
	if err != nil {
		return nil, err
	}

	for _, fragment := range fragments {
		name := filepath.Base(fragment)

		ts, err := template.New(name).Funcs(functions).ParseFiles(fragment)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
