//go:build prod
// +build prod

package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed "static"
var embeddedFiles embed.FS

func CreateFileServer() http.Handler {
	fsys, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		panic(err)
	}

	return http.FileServer(http.FS(fsys))
}
