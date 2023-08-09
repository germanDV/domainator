package events

import (
	"domainator/internal/httphelp"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AttachRoutes attaches the routes for the events controller.
func AttachRoutes(mux *httprouter.Router, controller *Controller) {
	mux.Handler(http.MethodPost, "/events", httphelp.Base.ThenFunc(controller.Save))
}
