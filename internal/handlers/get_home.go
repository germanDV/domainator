package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/templates"
)

func GetHome(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)
		certificates, err := certsService.GetAll(certs.GetAllCertsReq{UserID: userID})
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
