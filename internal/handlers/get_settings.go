package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/users"
)

func GetSettings(userService users.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDstr := cntxt.GetUserID(r)
		userID, err := common.ParseID(userIDstr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		u, err := userService.GetByID(r.Context(), users.GetByIDReq{UserID: userID})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		c := Layout(Settings(u.WebhookURL.String()), "Domainator | Settings")
		SendTempl(w, r, c)
	}
}
