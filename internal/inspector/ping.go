package inspector

import (
	"context"
	"domainator/internal/services"
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
		i.errorLog.Println(err)
	}

	for _, s := range settings {
		go i.pingDomain(s)
	}
}

// pingDomain pings a domain and saves the result to the database.
func (i Inspector) pingDomain(s *services.PingSettings) {
	i.infoLog.Printf("Pinging %q\n", s.Domain)

	start := time.Now()
	resp, err := http.Get(s.Domain)
	if err != nil {
		i.errorLog.Printf("Error pinging %q: %s\n", s.Domain, err.Error())
		return
	}

	// Read and close body even if we don't care about it,
	// to avoid leaking connections and keeping too many file descriptors.
	_, _ = io.ReadAll(resp.Body)
	defer resp.Body.Close()

	err = i.pinger.SavePingCheck(context.Background(), &services.Ping{
		ID:         uuid.New(),
		SettingsID: s.ID,
		TookMs:     int(time.Since(start).Milliseconds()),
		RespStatus: resp.StatusCode,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		i.errorLog.Printf("Error saving ping to db (%q): %s\n", s.Domain, err.Error())
	}
}
