package services

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pinger is an interface that the ping service must implement
type Pinger interface {
	Validate(payload Validatable) bool
	SaveSettings(ctx context.Context, payload *PingCreate) (uuid.UUID, error)
	GetSummary(ctx context.Context, userID uuid.UUID) ([]*PingSummary, error)
	GetSettingsByID(ctx context.Context, id uuid.UUID) (*PingSettings, error)
	GetChecksByID(ctx context.Context, id uuid.UUID) ([]*Ping, error)
	DeleteSettingsByID(ctx context.Context, id uuid.UUID) error
}

// PingService is a service that implements the Pinger interface
type PingService struct {
	Validator *validator.Validate
	DB        *pgxpool.Pool
}

// Ping represents a check to a domain
type Ping struct {
	ID         uuid.UUID
	SettingsID uuid.UUID
	RespStatus int
	TookMs     int
	CreatedAt  time.Time
}

// Validate validates a struct
func (ps *PingService) Validate(payload Validatable) bool {
	return payload.Validate(ps.Validator)
}

// GetSummary returns all ping settings for a user with data about the latest ping
func (ps *PingService) GetSummary(ctx context.Context, userID uuid.UUID) ([]*PingSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		with l as (
			select settings_id, max(created_at) as latest from pings group by settings_id
		)
		select
			ps.id,
			ps.domain,
			ps.success_code,
			coalesce(p.resp_status, 0) as resp_status,
			max(coalesce(p.created_at, '0001-01-01')) as last_check
		from ping_settings ps
		left outer join l on l.settings_id = ps.id
		left outer join pings p on
			p.settings_id = ps.id
			and p.created_at = l.latest
		group by ps.id, p.resp_status;
	`

	rows, err := ps.DB.Query(ctx, q)
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

	q := `insert into ping_settings (id, domain, success_code) values ($1, $2, $3)`
	_, err := ps.DB.Exec(ctx, q, newID.String(), payload.Domain, payload.SuccessCode)
	if err != nil {
		return uuid.Nil, err
	}

	return newID, nil
}

// GetSettingsByID returns a ping settings by its ID
func (ps *PingService) GetSettingsByID(ctx context.Context, id uuid.UUID) (*PingSettings, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, domain, success_code from ping_settings where id = $1`
	row := ps.DB.QueryRow(ctx, q, id.String())

	p := &PingSettings{}

	err := row.Scan(&p.ID, &p.Domain, &p.SuccessCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return p, nil
}

// GetChecksByID returns a ping by its ID
func (ps *PingService) GetChecksByID(ctx context.Context, id uuid.UUID) ([]*Ping, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sql := `select id, settings_id, resp_status, took_ms, created_at
		from pings where settings_id = $1
		order by created_at desc`

	rows, err := ps.DB.Query(ctx, sql, id.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pings := []*Ping{}

	for rows.Next() {

		p := &Ping{}
		err := rows.Scan(&p.ID, &p.SettingsID, &p.RespStatus, &p.TookMs, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		pings = append(pings, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return pings, nil
}

// DeleteSettingsByID deletes a ping settings and all its checks
func (ps *PingService) DeleteSettingsByID(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := ps.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "delete from pings where settings_id = $1", id.String())
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "delete from ping_settings where id = $1", id.String())
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
