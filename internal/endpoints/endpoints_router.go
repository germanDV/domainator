package endpoints

import (
	"domainator/internal/httphelp"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AttachRoutes attaches the routes for the endpoits controller.
func AttachRoutes(mux *httprouter.Router, controller *Controller) {
	mux.Handler(http.MethodGet, "/endpoints", httphelp.Protected.ThenFunc(controller.GetSummary))
	mux.Handler(http.MethodGet, "/endpoints-new", httphelp.Protected.ThenFunc(controller.EndpointNewForm))
	mux.Handler(http.MethodPost, "/endpoints-new", httphelp.Protected.ThenFunc(controller.CreateEndpoint))
	mux.Handler(http.MethodGet, "/endpoints/:id", httphelp.Protected.ThenFunc(controller.GetEndpoint))
	mux.Handler(http.MethodDelete, "/endpoints/:id", httphelp.Protected.ThenFunc(controller.DeleteByID))
}
