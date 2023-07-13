// Package inspector contains the Inspector type, which performs background tasks at a given interval.
package inspector

import (
	"domainator/internal/config"
	"domainator/internal/logger"
	"domainator/internal/notifier"
	"domainator/internal/services"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Inspector performs background tasks at a given interval.
type Inspector struct {
	DB               *pgxpool.Pool
	pinger           services.Pinger
	pingTickInterval time.Duration
	cleanInterval    time.Duration
	cleanMaxAge      time.Duration
	logit            *logger.Logit
	FailsCh          chan FailedPing
	Mailer           notifier.Notifier
	Slacker          notifier.Notifier
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
func New(db *pgxpool.Pool, pinger services.Pinger, logit *logger.Logit) Inspector {
	return Inspector{
		DB:               db,
		pinger:           pinger,
		pingTickInterval: config.GetDuration("PING_TICK_INTERVAL"),
		cleanInterval:    config.GetDuration("CLEAN_INTERVAL"),
		cleanMaxAge:      config.GetDuration("CLEAN_MAX_AGE"),
		logit:            logit,
		FailsCh:          make(chan FailedPing),
		Mailer:           notifier.NewMailer(),
		Slacker:          notifier.NewSlacker(),
	}
}

// Start kicks off the background tasks in a goroutine.
func (i Inspector) Start() {
	go i.startPingLoop()
	go i.startCleanLoop()
}
