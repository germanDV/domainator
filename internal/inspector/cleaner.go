package inspector

import (
	"context"
	"domainator/internal/logger"
	"fmt"
	"time"
)

// startCleanLoop sets an interval to delete old data from the db
// (it immediately performs a 'clean' and then sets the interval).
func (i Inspector) startCleanLoop() {
	ticker := time.NewTicker(i.cleanInterval)
	defer ticker.Stop()

	i.cleanHealthchecks()
	i.cleanCertchecks()
	for range ticker.C {
		i.cleanHealthchecks()
		i.cleanCertchecks()
	}
}

// cleanHealthchecks deletes old Healthchecks from the db.
func (i Inspector) cleanHealthchecks() {
	deleted, err := i.endpointsRepo.DeleteOldHealthchecks(context.Background(), i.cleanMaxAge)
	if err != nil {
		logger.Writer.Error(fmt.Sprintf("Error cleaning healthchecks: %v", err))
	} else {
		logger.Writer.Info(fmt.Sprintf("Removed %d healthchecks", deleted))
	}
}

func (i Inspector) cleanCertchecks() {
	deleted, err := i.certsRepo.DeleteOldChecks(context.Background(), i.cleanMaxAge)
	if err != nil {
		logger.Writer.Error(fmt.Sprintf("Error cleaning certchecks: %v", err))
	} else {
		logger.Writer.Info(fmt.Sprintf("Removed %d certchecks", deleted))
	}
}
