// Package inspector contains the Inspector type, which performs background tasks at a given interval.
package inspector

import (
	"domainator/internal/config"
	"domainator/internal/services"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Inspector performs background tasks at a given interval.
type Inspector struct {
	DB               *pgxpool.Pool
	pinger           services.Pinger
	pingTickInterval time.Duration
	cleanInterval    time.Duration
	cleanMaxAge      time.Duration
	errorLog         *log.Logger
	infoLog          *log.Logger
}

// New creates a new Inspector.
func New(db *pgxpool.Pool, pinger services.Pinger, errorLog, infoLog *log.Logger) Inspector {
	return Inspector{
		DB:               db,
		pinger:           pinger,
		pingTickInterval: config.GetDuration("PING_TICK_INTERVAL"),
		cleanInterval:    config.GetDuration("CLEAN_INTERVAL"),
		cleanMaxAge:      config.GetDuration("CLEAN_MAX_AGE"),
		errorLog:         errorLog,
		infoLog:          infoLog,
	}
}

// Start kicks off the background tasks in a goroutine.
func (i Inspector) Start() {
	go i.startPingLoop()
	go i.startCleanLoop()
}
