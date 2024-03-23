package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
)

func RegisterDomain(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)
		domain := r.FormValue("domain")

		req := RegisterCertReq{Domain: domain, UserID: userID}
		parsedReq, err := req.Parse()
		if err != nil {
			w.WriteHeader(400)
			e := RegisterDomainError(err.Error()).Render(r.Context(), w)
			if e != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		cert, err := certsService.Save(r.Context(), parsedReq)
		if err != nil {
			w.WriteHeader(400)
			e := RegisterDomainError(err.Error()).Render(r.Context(), w)
			if e != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		err = CertRow(serviceToTransportAdapter(cert)).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}
