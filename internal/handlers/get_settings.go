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

		c := Settings(u.WebhookURL.String())
		err = Layout(c, "Domainator | Settings", true).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}
