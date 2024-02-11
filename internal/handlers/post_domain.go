package handlers

import (
	"net/http"
	"time"

	"github.com/germandv/domainator/internal/domains/certs"
	"github.com/germandv/domainator/internal/templates"
)

func RegisterDomain(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domain := r.FormValue("domain")
		resp, err := certsService.RegisterCert(certs.RegisterCertReq{Domain: domain})
		if err != nil {
			w.WriteHeader(400)
			e := templates.RegisterDomainError(err.Error()).Render(r.Context(), w)
			if e != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		// TODO: remove after testing
		time.Sleep(1 * time.Second)

		err = templates.RegisterDomainSuccess(resp.ID, domain).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}
