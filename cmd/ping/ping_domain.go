package main

import (
	"context"
	"domainator/internal/services"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func doPings() {
	settings, err := pinger.GetSettings(context.Background())
	if err != nil {
		log.Println(err)
	}

	for _, s := range settings {
		go pingDomain(s)
	}
}

func pingDomain(s *services.PingSettings) {
	log.Printf("Pinging %q\n", s.Domain)

	start := time.Now()
	resp, err := http.Get(s.Domain)
	if err != nil {
		log.Printf("Error pinging %q: %s\n", s.Domain, err.Error())
		return
	}

	// Read and close body even if we don't care about it,
	// to avoid leaking connections and keeping too many file descriptors.
	_, _ = io.ReadAll(resp.Body)
	defer resp.Body.Close()

	err = pinger.SavePingCheck(context.Background(), &services.Ping{
		ID:         uuid.New(),
		SettingsID: s.ID,
		TookMs:     int(time.Since(start).Milliseconds()),
		RespStatus: resp.StatusCode,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		log.Printf("Error saving ping to db (%q): %s\n", s.Domain, err.Error())
	}
}
