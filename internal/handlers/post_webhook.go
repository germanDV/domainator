package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/users"
)

func SetWebhookURL(userService users.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDstr := cntxt.GetUserID(r)
		userID, err := common.ParseID(userIDstr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		urlInput := r.FormValue("webhook_url")

		url, err := common.ParseURL(urlInput)
		if err != nil {
			w.WriteHeader(400)
			e := WebhookForm(false, err.Error(), url.String()).Render(r.Context(), w)
			if e != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		req := users.SetWebhookReq{
			UserID: userID,
			URL:    url,
		}

		err = userService.SetWebhookURL(r.Context(), req)
		if err != nil {
			w.WriteHeader(400)
			e := WebhookForm(false, err.Error(), url.String()).Render(r.Context(), w)
			if e != nil {
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(200)
		e := WebhookForm(true, "", url.String()).Render(r.Context(), w)
		if e != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}
