package inspector

import (
	"context"
	"domainator/internal/bg"
	"domainator/internal/endpoints"
	"domainator/internal/logger"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// startHealthcheckLoop starts a loop that gets Endpoints from the db and pings them at a set interval.
func (i Inspector) startHealthcheckLoop() {
	ticker := time.NewTicker(i.healthcheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		i.doHealthchecks()
	}
}

// doHealthchecks gets all Endpoints from the database and fires off a goroutine to check each one.
func (i Inspector) doHealthchecks() {
	endpoints, err := i.endpointsRepo.GetAll(context.Background())
	if err != nil {
		logger.Writer.Error(err)
		return
	}

	// TODO: implement a semaphore to limit the number of concurrent requests

	for _, e := range endpoints {
		ee := e
		bg.Run(func() { i.ping(ee) })
	}
}

// ping makes a Healthcheck and saves the result to the database.
func (i Inspector) ping(s *endpoints.Endpoint) {
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
	err := i.endpointsRepo.SaveHealthcheck(context.Background(), &endpoints.Healthcheck{
		ID:         checkID,
		EndpointID: s.ID,
		TookMs:     int(time.Since(start).Milliseconds()),
		RespStatus: status,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		logger.Writer.Error(fmt.Sprintf("Error saving healthcheck to db (%q): %s", s.Domain, err.Error()))
		return
	}

	if status != s.SuccessCode {
		fp := FailedHealthcheck{
			EndpointID:   s.ID,
			CheckID:      checkID,
			URL:          s.Domain,
			ExpectedCode: s.SuccessCode,
			ActualCode:   status,
			Time:         start,
		}
		i.failsCh <- fp
	}
}
