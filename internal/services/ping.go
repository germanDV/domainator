package services

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pinger is an interface that the ping service must implement
type Pinger interface {
	Validate(payload Validatable) bool
	SaveSettings(ctx context.Context, payload *PingCreate) (uuid.UUID, error)
	GetSummary(ctx context.Context, userID uuid.UUID) ([]*PingSummary, error)
}

// PingService is a service that implements the Pinger interface
type PingService struct {
	Validator *validator.Validate
	DB        *pgxpool.Pool
}

// Validate validates a struct
func (ps *PingService) Validate(payload Validatable) bool {
	return payload.Validate(ps.Validator)
}

// GetSummary returns all ping settings for a user with data about the latest ping
func (ps *PingService) GetSummary(ctx context.Context, userID uuid.UUID) ([]*PingSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sql := `select id, domain from ping_settings`

	rows, err := ps.DB.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []*PingSummary{}
	for rows.Next() {
		s := &PingSummary{}
		err = rows.Scan(&s.ID, &s.Domain)
		if err != nil {
			return nil, err
		}
		s.Status = "healthy"
		s.LastCheck = time.Now()
		summaries = append(summaries, s)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return summaries, nil
}

// SaveSettings saves the settings for a domain to be pinged
func (ps *PingService) SaveSettings(ctx context.Context, payload *PingCreate) (uuid.UUID, error) {
	newID := uuid.New()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sql := `insert into ping_settings (id, domain, success_code) values ($1, $2, $3)`
	_, err := ps.DB.Exec(ctx, sql, newID.String(), payload.Domain, payload.SuccessCode)
	if err != nil {
		return uuid.Nil, err
	}

	return newID, nil
}
