package services

import (
	"time"

	"github.com/google/uuid"
)

// PingSettings represents a domain to ping with its settings
type PingSettings struct {
	ID          uuid.UUID
	Domain      string
	SuccessCode int
	CreatedAt   time.Time
}
