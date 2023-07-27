// Package tmpl provides functionality for rendering HTML templates.
package tmpl

import (
	"bytes"
	"domainator/internal/config"
	"domainator/internal/httphelp"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	pagesCache     = newPagesCache()
	fragmentsCache = newFragmentsCache()
)

// RenderPage is a helper that renders the templates with data, and handles any errors.
// It first renders the template into a buffer to check for errors, then writes the buffer to the http.ResponseWriter.
func RenderPage(w http.ResponseWriter, status int, page string, data *map[string]any) {
	ts, ok := pagesCache[page]
	if !ok {
		httphelp.ServerError(w, fmt.Errorf("The template %s does not exist", page))
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

// RenderFragment is a helper that renders an HTML fragment with data, and handles any errors.
func RenderFragment(w http.ResponseWriter, fragment string, data *map[string]any) {
	ts, ok := fragmentsCache[fragment]
	if !ok {
		httphelp.ServerError(w, fmt.Errorf("The fragment %s does not exist", fragment))
		return
	}

	buf := new(bytes.Buffer)
	tmplName := strings.TrimSuffix(fragment, ".html.tmpl")
	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	buf.WriteTo(w)
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// newPagesCache parses HTML templates and caches them.
// It panics if there is an error parsing the templates.
func newPagesCache() map[string]*template.Template {
	cache := map[string]*template.Template{}
	rootPath := config.GetRootPath()

	pages, err := filepath.Glob(rootPath + "/ui/html/pages/*.html.tmpl")
	if err != nil {
		panic(err)
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.ParseFiles(rootPath + "/ui/html/base.html.tmpl")
		if err != nil {
			panic(err)
		}

		ts, err = ts.ParseGlob(rootPath + "/ui/html/partials/*.html.tmpl")
		if err != nil {
			panic(err)
		}

		ts, err = ts.ParseGlob(rootPath + "/ui/html/fragments/*.html.tmpl")
		if err != nil {
			panic(err)
		}

		ts, err = ts.Funcs(functions).ParseFiles(page)
		if err != nil {
			panic(err)
		}

		cache[name] = ts
	}

	return cache
}

// newFragmentsCache parses HTML templates and caches them.
// It panics if there is an error parsing the templates.
func newFragmentsCache() map[string]*template.Template {
	cache := map[string]*template.Template{}
	rootPath := config.GetRootPath()

	fragments, err := filepath.Glob(rootPath + "/ui/html/fragments/*.html.tmpl")
	if err != nil {
		panic(err)
	}

	for _, fragment := range fragments {
		name := filepath.Base(fragment)
		ts, err := template.New(name).Funcs(functions).ParseFiles(fragment)
		if err != nil {
			panic(err)
		}
		cache[name] = ts
	}

	return cache
}

// BaseData is a helper that returns the basic common data for the templates.
func BaseData(r *http.Request) map[string]any {
	data := map[string]any{}
	data["Year"] = time.Now().Year()

	userID, ok := r.Context().Value(httphelp.UserIDContextKey).(uuid.UUID)
	if ok {
		data["User"] = map[string]any{
			"ID": userID.String(),
		}
	}

	return data
}
