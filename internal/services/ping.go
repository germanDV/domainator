package services

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pinger is an interface that the ping service must implement
type Pinger interface {
	SaveSettings(*PingCreate) (uuid.UUID, error)
	GetSummary(userID uuid.UUID) ([]*PingSummary, error)
	Validate(payload Validatable) bool
}

var dummyDB = []*PingSettings{
	{ID: uuid.New(), Domain: "debian.org", SuccessCode: 200, CreatedAt: time.Now()},
	{ID: uuid.New(), Domain: "wikipedia.com", SuccessCode: 200, CreatedAt: time.Now()},
	{ID: uuid.New(), Domain: "duckduckgo.com", SuccessCode: 201, CreatedAt: time.Now()},
}

// PingSettings is the settings for a domain to be pinged
type PingSettings struct {
	ID          uuid.UUID
	Domain      string
	SuccessCode int
	CreatedAt   time.Time
}

// PingSummary returns data about a ping settings and the latest check
type PingSummary struct {
	ID        uuid.UUID
	Domain    string
	Status    string // 'healthy' or 'unhealthy'
	LastCheck time.Time
}

// PingService is a ping service
type PingService struct {
	Validator *validator.Validate
	DB        *pgxpool.Pool
}

// Validate validates a struct
func (ps *PingService) Validate(payload Validatable) bool {
	return payload.Validate(ps.Validator)
}

// GetSummary returns all ping settings for a user with data about the latest ping
func (ps *PingService) GetSummary(userID uuid.UUID) ([]*PingSummary, error) {
	summaries := []*PingSummary{}
	for _, s := range dummyDB {
		summaries = append(summaries, &PingSummary{
			ID:        s.ID,
			Domain:    s.Domain,
			Status:    "healthy",
			LastCheck: time.Now(),
		})
	}
	return summaries, nil
}

// SaveSettings saves the settings for a domain to be pinged
func (ps *PingService) SaveSettings(payload *PingCreate) (uuid.UUID, error) {
	newID := uuid.New()
	dummyDB = append(dummyDB, &PingSettings{
		ID:          newID,
		Domain:      payload.Domain,
		SuccessCode: payload.SuccessCode,
		CreatedAt:   time.Now().UTC(),
	})
	return newID, errors.New("not implemented")
}
