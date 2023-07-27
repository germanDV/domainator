package inspector

import (
	"context"
	"domainator/internal/logger"
	"fmt"
	"time"
)

// startCleanLoop sets an interval to delete old data from the db
// (it immediately performs a 'clean' and then sets the interval)
func (i Inspector) startCleanLoop() {
	ticker := time.NewTicker(i.cleanInterval)
	quit := make(chan struct{})
	defer close(quit)

	i.cleanPings()

	for {
		select {
		case <-ticker.C:
			i.cleanPings()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// cleanPings deletes old ping checks from the db
func (i Inspector) cleanPings() {
	deleted, err := i.pingsRepo.DeleteOldPings(context.Background(), i.cleanMaxAge)
	if err != nil {
		logger.Writer.Error(fmt.Sprintf("Error cleaning pings: %v", err))
	} else {
		logger.Writer.Info(fmt.Sprintf("Removed %d pings", deleted))
	}
}
