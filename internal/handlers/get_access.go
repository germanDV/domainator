package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/cntxt"
)

func GetAccess() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)
		if userID != "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		c := Layout(Login(), "Domainator | Login")
		SendTempl(w, r, c)
	}
}
