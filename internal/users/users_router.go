package users

import (
	"domainator/internal/httphelp"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// AttachRoutes attaches the routes for the users controller
func AttachRoutes(mux *httprouter.Router, controller *Controller) {
	mux.Handler(http.MethodGet, "/user/signup", httphelp.Base.ThenFunc(controller.SignupForm))
	mux.Handler(http.MethodPost, "/user/signup", httphelp.Base.ThenFunc(controller.Signup))
	mux.Handler(http.MethodGet, "/user/login", httphelp.Base.ThenFunc(controller.LoginForm))
	mux.Handler(http.MethodPost, "/user/login", httphelp.Base.ThenFunc(controller.Login))
	mux.Handler(http.MethodPost, "/user/logout", httphelp.Protected.ThenFunc(controller.Logout))
	mux.Handler(http.MethodGet, "/user/verify", httphelp.Base.ThenFunc(controller.VerifyForm))
	mux.Handler(http.MethodPost, "/user/verify", httphelp.Base.ThenFunc(controller.Verify))
	mux.Handler(http.MethodGet, "/settings", httphelp.Protected.ThenFunc(controller.GetSettings))
	mux.Handler(http.MethodPut, "/settings/email/:id", httphelp.Protected.ThenFunc(controller.UpsertEmailSetting))
	mux.Handler(http.MethodPut, "/settings/toggle/:id", httphelp.Protected.ThenFunc(controller.TogglePref))
}
