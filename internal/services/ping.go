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

	sql := `select
			ps.id,
			ps.domain,
			ps.success_code,
			coalesce(p.resp_status, 0) as resp_status,
			coalesce(p.created_at, '0001-01-01') as last_check
		from ping_settings ps
		left outer join pings p on p.settings_id = ps.id
		order by p.created_at desc
		limit 1;`

	rows, err := ps.DB.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []*PingSummary{}

	for rows.Next() {
		var pingID uuid.UUID
		var domain string
		var successCode int
		var respStatus int
		var lastCheck time.Time

		err = rows.Scan(&pingID, &domain, &successCode, &respStatus, &lastCheck)
		if err != nil {
			return nil, err
		}

		s := &PingSummary{}
		s.ID = pingID
		s.Domain = domain
		if respStatus == 0 {
			s.Status = "-"
		} else if respStatus == successCode {
			s.Status = "healthy"
		} else {
			s.Status = "unhealthy"
		}
		if !lastCheck.IsZero() {
			s.LastCheck = lastCheck
		}
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
