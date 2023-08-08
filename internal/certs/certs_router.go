package certs

import (
	"domainator/internal/httphelp"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AttachRoutes attaches the routes for the certs controller.
func AttachRoutes(mux *httprouter.Router, controller *Controller) {
	mux.Handler(http.MethodGet, "/certs", httphelp.Protected.ThenFunc(controller.GetSummary))
	mux.Handler(http.MethodGet, "/certs-new", httphelp.Protected.ThenFunc(controller.CertsNewForm))
	mux.Handler(http.MethodPost, "/certs-new", httphelp.Protected.ThenFunc(controller.SaveDomain))
	mux.Handler(http.MethodDelete, "/certs/:id", httphelp.Protected.ThenFunc(controller.DeleteByID))
}
