package inspector

import (
	"context"
	"fmt"
)

// cleanHealthchecks deletes old Healthchecks from the db.
func (i Inspector) cleanHealthchecks(doneCh chan<- struct{}) {
	deleted, err := i.endpointsRepo.DeleteOldHealthchecks(context.Background(), i.cleanMaxAge)
	if err != nil {
		i.logger.Error(fmt.Sprintf("Error cleaning healthchecks: %v", err))
	} else {
		i.logger.Info(fmt.Sprintf("Removed %d healthchecks", deleted))
	}
	doneCh <- struct{}{}
}

// cleanCertchecks deletes old Cert checks from the db.
func (i Inspector) cleanCertchecks(doneCh chan<- struct{}) {
	deleted, err := i.certsRepo.DeleteOldChecks(context.Background(), i.cleanMaxAge)
	if err != nil {
		i.logger.Error(fmt.Sprintf("Error cleaning certchecks: %v", err))
	} else {
		i.logger.Info(fmt.Sprintf("Removed %d certchecks", deleted))
	}
	doneCh <- struct{}{}
}
