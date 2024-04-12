package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
)

func GetHome(certsService certs.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)
		if userID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
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

		c := Index(transportCerts)
		err = Layout(c, "The Home Of The Domainator", true).Render(r.Context(), w)
    // TODO: make a utility for this because it's repeated every time a template is rendered.
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}
