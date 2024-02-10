package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/templates"
)

func GetHome(w http.ResponseWriter, r *http.Request) {
	c := templates.Index()
	err := templates.Layout(c, "The Home Of The Domainator").Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
