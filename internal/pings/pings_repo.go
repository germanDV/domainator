package pings

import (
	"context"
	"domainator/internal/validation"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo is an interface that the pings repository must implement.
type Repo interface {
	GetSummary(ctx context.Context, userID uuid.UUID) ([]*Summary, error)
	SaveSettings(ctx context.Context, userID uuid.UUID, payload *CreatePingReq) (uuid.UUID, error)
	Save(ctx context.Context, payload *Ping) error
	GetSettings(ctx context.Context) ([]*Settings, error)
	GetSettingsByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Settings, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]*Ping, error)
	DeleteSettingsByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	DeleteOldPings(ctx context.Context, age time.Duration) (int64, error)
}

// PostgresRepo is a repository that implements the Repo interface.
type PostgresRepo struct {
	DB *pgxpool.Pool
}

// NewPostgresRepo returns a new instance of a PostgresRepo.
func NewPostgresRepo(db *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{
		DB: db,
	}
}

// GetSummary returns all ping settings for a user with data about the latest ping.
func (pg *PostgresRepo) GetSummary(ctx context.Context, userID uuid.UUID) ([]*Summary, error) {
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
		where ps.user_id = $1
		group by ps.id, p.resp_status;
	`

	rows, err := pg.DB.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []*Summary{}

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

		s := &Summary{}
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

// SaveSettings saves the settings for a domain to be pinged.
func (pg *PostgresRepo) SaveSettings(ctx context.Context, userID uuid.UUID, payload *CreatePingReq) (uuid.UUID, error) {
	newID := uuid.New()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into ping_settings (id, domain, success_code, user_id) values ($1, $2, $3, $4)`
	_, err := pg.DB.Exec(ctx, q, newID.String(), payload.Domain, payload.SuccessCode, userID)
	if err != nil {
		return uuid.Nil, err
	}

	return newID, nil
}

// Save saves the result of pinging a domain
func (pg *PostgresRepo) Save(ctx context.Context, p *Ping) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into pings (id, settings_id, resp_status, took_ms) values ($1, $2, $3, $4)`
	_, err := pg.DB.Exec(ctx, q, p.ID.String(), p.SettingsID.String(), p.RespStatus, p.TookMs)
	if err != nil {
		return err
	}

	return nil
}

// GetSettings returns all ping settings for all users
func (pg *PostgresRepo) GetSettings(ctx context.Context) ([]*Settings, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, domain, success_code from ping_settings`
	rows, err := pg.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	settings := []*Settings{}
	for rows.Next() {
		p := &Settings{}
		err := rows.Scan(&p.ID, &p.Domain, &p.SuccessCode)
		if err != nil {
			return nil, err
		}
		settings = append(settings, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// GetSettingsByID returns a ping settings by its ID.
func (pg *PostgresRepo) GetSettingsByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Settings, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, domain, success_code from ping_settings where id = $1 and user_id = $2`
	row := pg.DB.QueryRow(ctx, q, id, userID)

	p := &Settings{}

	err := row.Scan(&p.ID, &p.Domain, &p.SuccessCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, validation.ErrNotFound
		}
		return nil, err
	}

	return p, nil
}

// GetByID returns a ping by its ID.
func (pg *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]*Ping, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		select p.id, p.settings_id, p.resp_status, p.took_ms, p.created_at
		from pings p
		inner join ping_settings ps
			on ps.id = p.settings_id
		where p.settings_id = $1
			and ps.user_id = $2
		order by p.created_at desc;
	`

	rows, err := pg.DB.Query(ctx, q, id, userID)
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

// DeleteSettingsByID deletes a ping settings and all its checks.
func (pg *PostgresRepo) DeleteSettingsByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := pg.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	resp, err := tx.Exec(ctx, "delete from ping_settings where id = $1 and user_id = $2", id, userID)
	if err != nil || resp.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "delete from pings where settings_id = $1", id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// DeleteOldPings deletes all ping checks older than the given age
func (pg *PostgresRepo) DeleteOldPings(ctx context.Context, age time.Duration) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := "delete from pings where created_at < $1"
	arg := time.Now().UTC().Add(-age).Format(time.DateTime)
	resp, err := pg.DB.Exec(ctx, q, arg)
	if err != nil {
		return 0, err
	}
	return resp.RowsAffected(), nil
}
