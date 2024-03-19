package handlers

import (
	"net/http"

	"github.com/germandv/domainator/internal/cntxt"
)

// TODO: render access page with GitHub OAuth
func GetAccess() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := cntxt.GetUserID(r)
		if userID != "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		w.Write([]byte("<h1>Beautiful Page with Sign In / Sign Up Options</h1>"))
	}
}
