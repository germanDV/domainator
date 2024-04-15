package handlers

import (
	"errors"
	"net/http"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
)

func UpdateDomain(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)

		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "No ID provided", http.StatusBadRequest)
			return
		}

		req := UpdateCertReq{UserID: userID, ID: id}
		parsedReq, err := req.Parse()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cert, err := certsService.Update(r.Context(), parsedReq)
		if err != nil {
			if errors.Is(err, certs.ErrNotFound) {
				http.Error(w, "Domain not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error updating domain", http.StatusInternalServerError)
			}
			return
		}

		c := CertRow(serviceToTransportAdapter(cert))
		SendTempl(w, r, c)
	}
}
