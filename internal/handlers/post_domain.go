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
			c := RegisterDomainError(err.Error())
			SendTemplWithStatus(http.StatusBadRequest, w, r, c)
			return
		}

		cert, err := certsService.Save(r.Context(), parsedReq)
		if err != nil {
			c := RegisterDomainError(err.Error())
			SendTemplWithStatus(http.StatusBadRequest, w, r, c)
			return
		}

		c := CertRow(serviceToTransportAdapter(cert))
		SendTempl(w, r, c)
	}
}
