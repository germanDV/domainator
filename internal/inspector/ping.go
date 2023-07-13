package inspector

import (
	"context"
	"domainator/internal/services"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// startPingLoop starts a loop that gets domains from the db and pings them at a set interval.
func (i Inspector) startPingLoop() {
	ticker := time.NewTicker(i.pingTickInterval)
	quit := make(chan struct{})
	defer close(quit)
	defer close(i.FailsCh)

	// TODO: remove, just to make it faster during development
	i.doPings()

	for {
		select {
		case <-ticker.C:
			i.doPings()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// doPings gets all ping settings from the database and fires off a goroutine to check each domain.
func (i Inspector) doPings() {
	settings, err := i.pinger.GetSettings(context.Background())
	if err != nil {
		i.logit.Error(err)
	}

	for _, s := range settings {
		go i.pingDomain(s)
	}
}

// pingDomain pings a domain and saves the result to the database.
func (i Inspector) pingDomain(s *services.PingSettings) {
	i.logit.Info(fmt.Sprintf("Pinging %q", s.Domain))

	start := time.Now()
	// TODO: use a custom http client with a timeout
	resp, err := http.Get(s.Domain)
	if err != nil {
		i.logit.Error(fmt.Sprintf("Error pinging %q: %s", s.Domain, err.Error()))
		return
	}

	// Read and close body even if we don't care about it,
	// to avoid leaking connections and keeping too many file descriptors.
	_, _ = io.ReadAll(resp.Body)
	defer resp.Body.Close()

	checkID := uuid.New()
	err = i.pinger.SavePingCheck(context.Background(), &services.Ping{
		ID:         checkID,
		SettingsID: s.ID,
		TookMs:     int(time.Since(start).Milliseconds()),
		RespStatus: resp.StatusCode,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		i.logit.Error(fmt.Sprintf("Error saving ping to db (%q): %s", s.Domain, err.Error()))
		return
	}

	if resp.StatusCode != s.SuccessCode {
		fp := FailedPing{
			SettingsID:   s.ID,
			CheckID:      checkID,
			URL:          s.Domain,
			ExpectedCode: s.SuccessCode,
			ActualCode:   resp.StatusCode,
			Time:         start,
		}
		i.FailsCh <- fp
	}
}
