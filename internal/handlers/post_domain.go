package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/domains/certs"
	"github.com/germandv/domainator/internal/templates"
)

func RegisterDomain(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.FormValue("domain")
		cert, err := certsService.RegisterCert(certs.RegisterCertReq{Domain: domain})
		if err != nil {
			w.WriteHeader(400)
			e := templates.RegisterDomainError(err.Error()).Render(r.Context(), w)
			if e != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		err = templates.RegisterDomainSuccess(cert).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}
