package handlers

import (
	"errors"
	"net/http"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
)

func DeleteDomain(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "No ID provided", http.StatusBadRequest)
			return
		}

		userID := cntxt.GetUserID(r)
		req := DeleteCertReq{UserID: userID, ID: id}
		parsedReq, err := req.Parse()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = certsService.Delete(r.Context(), parsedReq)
		if err != nil {
			if errors.Is(err, certs.ErrNotFound) {
				http.Error(w, "Domain not found", http.StatusNotFound)
			} else {
				http.Error(w, "Error deleting domain", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
