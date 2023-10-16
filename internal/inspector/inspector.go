// Package inspector contains the Inspector type, which performs background tasks at a given interval.
package inspector

import (
	"context"
	"domainator/internal/certs"
	"domainator/internal/certstatus"
	"domainator/internal/config"
	"domainator/internal/endpoints"
	"domainator/internal/notificators"
	"domainator/internal/notifier"
	"domainator/internal/users"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Inspector performs background tasks at a given interval.
type Inspector struct {
	usersRepo           users.Repo
	endpointsRepo       endpoints.Repo
	certsRepo           certs.Repo
	healthcheckInterval time.Duration
	certcheckInterval   time.Duration
	cleanInterval       time.Duration
	cleanMaxAge         time.Duration
	failsCh             chan FailedHealthcheck
	badCertsCh          chan BadCert
	mailer              notifier.Notifier
	slacker             notifier.Notifier
	httpclient          *http.Client
	dialer              *net.Dialer
	logger              *slog.Logger
}

// FailedHealthcheck represents a ping to an Endpoint that failed.
type FailedHealthcheck struct {
	EndpointID   uuid.UUID
	CheckID      uuid.UUID
	URL          string
	ExpectedCode int
	ActualCode   int
	Time         time.Time
}

// BadCert represents a certificate that failed a check or is about to expire.
type BadCert struct {
	CertID uuid.UUID
	Domain string
	Expiry time.Time
	Status certstatus.Status
	Time   time.Time
}

// New creates a new Inspector.
func New(db *pgxpool.Pool) Inspector {
	return Inspector{
		usersRepo:     users.NewPostgresRepo(db),
		endpointsRepo: endpoints.NewPostgresRepo(db),
		certsRepo:     certs.NewPostgresRepo(db),
		cleanMaxAge:   config.GetDuration("CLEAN_MAX_AGE"),
		failsCh:       make(chan FailedHealthcheck),
		badCertsCh:    make(chan BadCert),
		mailer:        notifier.NewMailer(),
		slacker:       notifier.NewSlacker(),
		httpclient:    &http.Client{Timeout: config.GetDuration("HEALTHCHECK_TIMEOUT")},
		dialer:        &net.Dialer{Timeout: config.GetDuration("HEALTHCHECK_TIMEOUT")},
	}
}

// Start kicks off the background tasks in a goroutine.
func (i Inspector) Start() {
	defer close(i.failsCh)

	doneCh := make(chan struct{})
	tasks := 4
	done := 0

	go i.doHealthChecks(doneCh)
	go i.doCertChecks(doneCh)
	go i.cleanHealthchecks(doneCh)
	go i.cleanCertchecks(doneCh)

	for {
		select {
		case fail := <-i.failsCh:
			i.handleFailedHealthcheck(fail)
		case badCert := <-i.badCertsCh:
			i.handleBadCert(badCert)
		case <-doneCh:
			done++
			if done == tasks {
				return
			}
		}
	}
}

func (i Inspector) handleFailedHealthcheck(fail FailedHealthcheck) {
	prefs, err := i.usersRepo.GetNotificationPrefsByEndpoint(context.Background(), fail.EndpointID)
	if err != nil {
		i.logger.Error(err.Error())
		return
	}

	if len(prefs) == 0 {
		i.logger.Info("User does not have any notification preferences set")
		return
	}

	for _, pref := range prefs {
		switch pref.Service {
		case notificators.Email:
			sub, body, err := parseFailedHealthcheckTemplate(fail)
			if err != nil {
				i.logger.Error(err.Error())
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
			i.logger.Info(fmt.Sprintf("Unknown notification type %q", pref.Service))
		}
	}
}

func parseFailedHealthcheckTemplate(fail FailedHealthcheck) (string, string, error) {
	return notifier.ParseTemplate("alert_healthcheck.html.tmpl", map[string]any{
		"URL":      fail.URL,
		"Expected": fail.ExpectedCode,
		"Actual":   fail.ActualCode,
		"Time":     fail.Time.UTC().Format("2006-01-02 15:04:05"),
	})
}

func (i Inspector) handleBadCert(badCert BadCert) {
	prefs, err := i.usersRepo.GetNotificationPrefsByCert(context.Background(), badCert.CertID)
	if err != nil {
		i.logger.Error(err.Error())
		return
	}

	if len(prefs) == 0 {
		i.logger.Info("User does not have any notification preferences set")
		return
	}

	for _, pref := range prefs {
		switch pref.Service {
		case notificators.Email:
			sub, body, err := parseBadCertTemplate(badCert)
			if err != nil {
				i.logger.Error(err.Error())
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
				Body: fmt.Sprintf("Trouble with certificate for domain %q: %s.", badCert.Domain, badCert.Status),
			})
		default:
			i.logger.Info(fmt.Sprintf("Unknown notification type %q", pref.Service))
		}
	}
}

func parseBadCertTemplate(badCert BadCert) (string, string, error) {
	return notifier.ParseTemplate("alert_cert.html.tmpl", map[string]any{
		"Domain": badCert.Domain,
		"Expiry": badCert.Expiry.UTC().Format("2006-01-02 15:04:05"),
		"Status": badCert.Status,
		"Time":   badCert.Time.UTC().Format("2006-01-02 15:04:05"),
	})
}
