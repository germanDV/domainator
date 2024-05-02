//go:build !prod
// +build !prod

package ui

import (
	"net/http"
)

func CreateFileServer() http.Handler {
	return http.FileServer(http.Dir("./ui/static"))
}
