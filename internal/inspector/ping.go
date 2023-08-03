package inspector

import (
	"context"
	"domainator/internal/bg"
	"domainator/internal/logger"
	"domainator/internal/pings"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// startPingLoop starts a loop that gets domains from the db and pings them at a set interval.
func (i Inspector) startPingLoop() {
	ticker := time.NewTicker(i.pingTickInterval)
	defer ticker.Stop()

	for range ticker.C {
		i.doPings()
	}
}

// doPings gets all ping settings from the database and fires off a goroutine to check each domain.
func (i Inspector) doPings() {
	settings, err := i.pingsRepo.GetSettings(context.Background())
	if err != nil {
		logger.Writer.Error(err)
	}

	for _, s := range settings {
		ss := s
		bg.Run(func() { i.pingDomain(ss) })
	}
}

// pingDomain pings a domain and saves the result to the database.
func (i Inspector) pingDomain(s *pings.Settings) {
	logger.Writer.Info(fmt.Sprintf("Pinging %q", s.Domain))

	start := time.Now()
	var status int

	req, err1 := http.NewRequest(http.MethodGet, s.Domain, nil)
	resp, err2 := i.httpclient.Do(req)
	if err1 != nil || err2 != nil {
		status = 523 // Unreachable
	} else {
		status = resp.StatusCode
		// Read and close body even if we don't care about it,
		// to avoid leaking connections and keeping too many file descriptors.
		_, _ = io.ReadAll(resp.Body)
		defer resp.Body.Close()
	}

	checkID := uuid.New()
	err := i.pingsRepo.Save(context.Background(), &pings.Ping{
		ID:         checkID,
		SettingsID: s.ID,
		TookMs:     int(time.Since(start).Milliseconds()),
		RespStatus: status,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		logger.Writer.Error(fmt.Sprintf("Error saving ping to db (%q): %s", s.Domain, err.Error()))
		return
	}

	if status != s.SuccessCode {
		fp := FailedPing{
			SettingsID:   s.ID,
			CheckID:      checkID,
			URL:          s.Domain,
			ExpectedCode: s.SuccessCode,
			ActualCode:   status,
			Time:         start,
		}
		i.failsCh <- fp
	}
}
