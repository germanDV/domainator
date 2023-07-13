package main

import (
	"context"
	"domainator/internal/inspector"
	"domainator/internal/notifier"
	"fmt"
)

func (app *application) startInspector() {
	app.inspector.Start()
	for fail := range app.inspector.FailsCh {
		app.handleFailedPing(fail)
	}
}

func (app *application) handleFailedPing(fail inspector.FailedPing) {
	prefs, err := app.userSvc.GetNotificationPreferencesBySettings(context.Background(), fail.SettingsID)
	if err != nil {
		app.logit.Error(err)
		return
	}

	if len(prefs) == 0 {
		app.logit.Info("User does not have any notification preferences set")
		return
	}

	for _, pref := range prefs {
		switch pref.Service {
		case "email":
			app.inspector.Mailer.Notify(notifier.Message{
				To:      pref.To,
				Subject: "Domainator: unhealthy domain!",
				// TODO: use an HTML template
				Body: fmt.Sprintf("<h2>Domain %q is unhealthy</h2><p>Want: %d, got: %d</p>", fail.URL, fail.ExpectedCode, fail.ActualCode),
			})
		case "slack":
			app.logit.Info("Sending slack message")
			app.inspector.Slacker.Notify(notifier.Message{
				To:      pref.To,
				Subject: "Domainator: unhealthy domain!",
				Body:    fmt.Sprintf("Domain %q is unhealthy. Want: %d, got: %d", fail.URL, fail.ExpectedCode, fail.ActualCode),
			})
		default:
			app.logit.Info(fmt.Sprintf("Unknown notification type %q", pref.Service))
		}
	}
}
