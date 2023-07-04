package services

import (
	"time"

	"github.com/google/uuid"
)

// PingSettings is the settings for a domain to be pinged
type PingSettings struct {
	ID          uuid.UUID
	Domain      string
	SuccessCode int
	CreatedAt   time.Time
}
