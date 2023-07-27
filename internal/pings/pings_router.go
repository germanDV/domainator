package pings

import (
	"domainator/internal/httphelp"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AttachRoutes attaches the routes for the pings controller
func AttachRoutes(mux *httprouter.Router, controller *Controller) {
	mux.Handler(http.MethodGet, "/pings", httphelp.Protected.ThenFunc(controller.GetSummary))
	mux.Handler(http.MethodGet, "/pings-new", httphelp.Protected.ThenFunc(controller.PingsNewForm))
	mux.Handler(http.MethodPost, "/pings-new", httphelp.Protected.ThenFunc(controller.CreatePing))
	mux.Handler(http.MethodGet, "/pings/:id", httphelp.Protected.ThenFunc(controller.GetPing))
	mux.Handler(http.MethodDelete, "/pings/:id", httphelp.Protected.ThenFunc(controller.DeleteByID))
}
