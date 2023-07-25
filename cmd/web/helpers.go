package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
)

// serverError helper writes an error message and stack trace to the errorLog,
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.logit.ErrorLog.Output(2, trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// clientError helper sends a specific status code and corresponding description to the user.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound is simply a convenience wrapper around clientError which sends a 404 to the user.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

// render is a helper that renders the templates with data, and handles any errors.
// It first renders the template into a buffer to check for errors, then writes the buffer to the http.ResponseWriter.
func (app *application) render(w http.ResponseWriter, status int, page string, data *map[string]any) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("The template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

// renderFragment is a helper that renders an HTML fragment with data, and handles any errors.
func (app *application) renderFragment(w http.ResponseWriter, fragment string, data *map[string]any) {
	ts, ok := app.fragmentCache[fragment]
	if !ok {
		err := fmt.Errorf("The fragment %s does not exist", fragment)
		app.serverError(w, err)
		return
	}

	buf := new(bytes.Buffer)
	tmplName := strings.TrimSuffix(fragment, ".html.tmpl")
	err := ts.ExecuteTemplate(buf, tmplName, data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	buf.WriteTo(w)
}

// decodeForm is a helper that decodes the form data from the request into the destination struct.
func (app *application) decodeForm(r *http.Request, dst any) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(dst, r.PostForm)
	if err != nil {
		return err
	}

	return nil
}

// initialTmplData is a helper that returns the basic common data for the templates.
func initialTmplData(r *http.Request) map[string]any {
	data := map[string]any{}
	data["Year"] = time.Now().Year()

	userID, ok := r.Context().Value(userIDContextKey).(uuid.UUID)
	if ok {
		data["User"] = map[string]any{
			"ID": userID.String(),
		}
	}

	return data
}

// GetUserIDFromCtx returns the user ID from the request context
func (app *application) GetUserIDFromCtx(w http.ResponseWriter, r *http.Request) uuid.UUID {
	userID, ok := r.Context().Value(userIDContextKey).(uuid.UUID)
	if !ok || userID == uuid.Nil || userID.String() == "" {
		return uuid.Nil
	}
	return userID
}

// find returns the first element in the slice that matches the predicate function.
func find[T any](slice []T, predicate func(T) bool) T {
	for _, v := range slice {
		if predicate(v) {
			return v
		}
	}

	empty := new(T)
	return *empty
}
