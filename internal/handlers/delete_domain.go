package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/certs"
)

func DeleteDomain(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "No ID provided", http.StatusBadRequest)
			return
		}

		err := certsService.Delete(id)
		if err != nil {
			http.Error(w, "Error deleting domain", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
