package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/domains/certs"
	"github.com/germandv/domainator/internal/templates"
)

func GetHome(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		certificates, err := certsService.GetAll()
		if err != nil {
			http.Error(w, "Error getting certificates", http.StatusInternalServerError)
			return
		}

		c := templates.Index(certificates)
		err = templates.Layout(c, "The Home Of The Domainator").Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}
