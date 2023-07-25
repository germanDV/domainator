package main

import (
	"domainator/internal/notificators"
	"domainator/internal/services"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func (app *application) settings(w http.ResponseWriter, r *http.Request) {
	userID := app.GetUserIDFromCtx(w, r)
	if userID == uuid.Nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	user, err := app.userSvc.GetByID(r.Context(), userID)
	if err != nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	prefs, err := app.userSvc.GetNotificationPreferencesByUserID(r.Context(), userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	emailSettings := find[services.NotificationPreference](prefs, func(p services.NotificationPreference) bool {
		return p.Service == notificators.Email
	})
	slackSettings := find[services.NotificationPreference](prefs, func(p services.NotificationPreference) bool {
		return p.Service == notificators.Slack
	})

	templateData := initialTmplData(r)
	templateData["User"] = map[string]string{"Email": user.Email}
	templateData["Prefs"] = map[string]services.NotificationPreference{
		"Email": emailSettings,
		"Slack": slackSettings,
	}

	app.render(w, http.StatusOK, "settings.html.tmpl", &templateData)
}

func (app *application) settingsUpsertEmail(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userID := app.GetUserIDFromCtx(w, r)
	if userID == uuid.Nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	var payload services.EmailUpdate
	app.decodeForm(r, &payload)

	ok := app.userSvc.Validate(&payload)
	if !ok {
		data := map[string]any{
			"Prefs": map[string]any{
				"Email": map[string]any{
					"ID": id,
					"To": payload.Email,
				},
			},
			"Error": payload.Errors["Email"],
		}
		app.renderFragment(w, "settings_email_input.html.tmpl", &data)
		return
	}

	var pref *services.NotificationPreference
	var e error
	if id == 0 {
		pref, err = app.userSvc.CreateEmailNotification(r.Context(), userID, payload.Email)
	} else {
		pref, err = app.userSvc.UpdateEmailNotification(r.Context(), id, payload.Email)
	}
	if e != nil {
		app.serverError(w, err)
		return
	}

	emailSettings := map[string]any{"ID": pref.ID, "To": pref.To}
	data := map[string]any{
		"Prefs": map[string]any{
			"Email": emailSettings,
		},
	}

	app.renderFragment(w, "settings_email_input.html.tmpl", &data)
}
