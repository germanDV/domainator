package handlers

import (
	"net/http"
)

func GetAccess() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("<h1>Beautiful Page with Sign In / Sign Up Options</h1>"))
	}
}
