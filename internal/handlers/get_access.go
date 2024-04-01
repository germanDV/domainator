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

		c := Login()
		err := Layout(c, "Domainator | Login", false).Render(r.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
	}
}
