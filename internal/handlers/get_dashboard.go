package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
)

func GetDashboard(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)
		if userID == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		req := GetAllCertsReq{UserID: userID}
		parsedReq, err := req.Parse()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		certificates, err := certsService.GetAll(r.Context(), parsedReq)
		if err != nil {
			http.Error(w, "Error getting certificates", http.StatusInternalServerError)
			return
		}

		transportCerts := make([]TransportCert, len(certificates))
		for i, cert := range certificates {
			transportCerts[i] = serviceToTransportAdapter(cert)
		}

		c := Layout(Dashboard(transportCerts), "Domainator | Dashboard")
		SendTempl(w, r, c)
	}
}
