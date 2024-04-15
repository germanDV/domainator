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
			c := WebhookForm(false, err.Error(), url.String())
			SendTemplWithStatus(http.StatusBadRequest, w, r, c)
			return
		}

		req := users.SetWebhookReq{
			UserID: userID,
			URL:    url,
		}

		err = userService.SetWebhookURL(r.Context(), req)
		if err != nil {
			c := WebhookForm(false, err.Error(), url.String())
			SendTemplWithStatus(http.StatusBadRequest, w, r, c)
			return
		}

		c := WebhookForm(true, "", url.String())
		SendTempl(w, r, c)
	}
}
