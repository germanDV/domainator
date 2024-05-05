package handlers

import (
	"net/http"
)

func GetLanding() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := Layout(Landing(), "Domainator")
		SendTempl(w, r, c)
	}
}
