package services

import (
	"time"

	"github.com/google/uuid"
)

// PingSummary returns data about a ping settings and the latest check
type PingSummary struct {
	ID        uuid.UUID
	Domain    string
	Status    string // 'healthy' or 'unhealthy'
	LastCheck time.Time
}
