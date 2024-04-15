package handlers

import (
	"net/http"

	"github.com/a-h/templ"
)

// SendTempl renders a template and sends it to the client,
// if there's an error, it sends a 500.
func SendTempl(w http.ResponseWriter, r *http.Request, c templ.Component) {
	err := c.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// SendTemplWithSatus sends a template with a status code instead of defaulting to 200.
func SendTemplWithStatus(s int, w http.ResponseWriter, r *http.Request, c templ.Component) {
	w.WriteHeader(s)
	SendTempl(w, r, c)
}
