// Package inspector contains the Inspector type, which performs background tasks at a given interval.
package inspector

import (
	"context"
	"domainator/internal/bg"
	"domainator/internal/config"
	"domainator/internal/logger"
	"domainator/internal/notificators"
	"domainator/internal/notifier"
	"domainator/internal/pings"
	"domainator/internal/users"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Inspector performs background tasks at a given interval.
type Inspector struct {
	pingsRepo        pings.Repo
	usersRepo        users.Repo
	pingTickInterval time.Duration
	cleanInterval    time.Duration
	cleanMaxAge      time.Duration
	failsCh          chan FailedPing
	mailer           notifier.Notifier
	slacker          notifier.Notifier
	httpclient       *http.Client
}

// FailedPing is a ping to a domain/url that failed.
type FailedPing struct {
	SettingsID   uuid.UUID
	CheckID      uuid.UUID
	URL          string
	ExpectedCode int
	ActualCode   int
	Time         time.Time
}

// New creates a new Inspector.
func New(db *pgxpool.Pool) Inspector {
	return Inspector{
		pingsRepo:        pings.NewPostgresRepo(db),
		usersRepo:        users.NewPostgresRepo(db),
		pingTickInterval: config.GetDuration("PING_TICK_INTERVAL"),
		cleanInterval:    config.GetDuration("CLEAN_INTERVAL"),
		cleanMaxAge:      config.GetDuration("CLEAN_MAX_AGE"),
		failsCh:          make(chan FailedPing),
		mailer:           notifier.NewMailer(),
		slacker:          notifier.NewSlacker(),
		httpclient:       &http.Client{Timeout: config.GetDuration("PING_TIMEOUT")},
	}
}

// Start kicks off the background tasks in a goroutine.
func (i Inspector) Start() {
	bg.Run(i.startPingLoop)
	bg.Run(i.startCleanLoop)

	for fail := range i.failsCh {
		i.handleFailedPing(fail)
	}
}

func (i Inspector) handleFailedPing(fail FailedPing) {
	prefs, err := i.usersRepo.GetNotificationPrefsBySettings(context.Background(), fail.SettingsID)
	if err != nil {
		logger.Writer.Error(err)
		return
	}

	if len(prefs) == 0 {
		logger.Writer.Info("User does not have any notification preferences set")
		return
	}

	for _, pref := range prefs {
		switch pref.Service {
		case notificators.Email:
			sub, body, err := parseFailedPingTemplate(fail)
			if err != nil {
				logger.Writer.Error(err)
				continue
			}
			i.mailer.Notify(notifier.Message{
				To:      pref.To,
				Subject: sub,
				Body:    body,
			})
		case notificators.Slack:
			i.slacker.Notify(notifier.Message{
				To:   pref.To,
				Body: fmt.Sprintf("Domain %q is unhealthy. Want: %d, got: %d", fail.URL, fail.ExpectedCode, fail.ActualCode),
			})
		default:
			logger.Writer.Info(fmt.Sprintf("Unknown notification type %q", pref.Service))
		}
	}
}

// parseFailedPingTemplate parses the corresponding template and returns the subject and body
func parseFailedPingTemplate(fail FailedPing) (string, string, error) {
	return notifier.ParseTemplate("alert_ping.html.tmpl", map[string]any{
		"URL":      fail.URL,
		"Expected": fail.ExpectedCode,
		"Actual":   fail.ActualCode,
		"Time":     fail.Time.UTC().Format("2006-01-02 15:04:05"),
	})
}
