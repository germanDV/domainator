package endpoints

import (
	"context"
	"domainator/internal/validation"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo is an interface that the endpoints repository must implement.
type Repo interface {
	Save(ctx context.Context, userID uuid.UUID, payload *CreateEndpointReq) (uuid.UUID, error)
	SaveHealthcheck(ctx context.Context, payload *Healthcheck) error
	GetSummary(ctx context.Context, userID uuid.UUID) ([]*Summary, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Endpoint, error)
	GetHealthcheckByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]*Healthcheck, error)
	GetAll(ctx context.Context) ([]*Endpoint, error)
	Count(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	DeleteOldHealthchecks(ctx context.Context, age time.Duration) (int64, error)
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

// Save creates a new Endpoint and saves it to the database.
func (pg *PostgresRepo) Save(ctx context.Context, userID uuid.UUID, payload *CreateEndpointReq) (uuid.UUID, error) {
	newID := uuid.New()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into endpoints (id, domain, success_code, user_id) values ($1, $2, $3, $4)`
	_, err := pg.DB.Exec(ctx, q, newID.String(), payload.Domain, payload.SuccessCode, userID)
	if err != nil {
		return uuid.Nil, err
	}

	return newID, nil
}

// SaveHealthcheck saves the result of pinging a domain.
func (pg *PostgresRepo) SaveHealthcheck(ctx context.Context, p *Healthcheck) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into healthchecks (id, endpoint_id, resp_status, took_ms) values ($1, $2, $3, $4)`
	_, err := pg.DB.Exec(ctx, q, p.ID, p.EndpointID, p.RespStatus, p.TookMs)
	if err != nil {
		return err
	}

	return nil
}

// GetSummary returns all Endpoints for a user with their latest Healthchecks.
func (pg *PostgresRepo) GetSummary(ctx context.Context, userID uuid.UUID) ([]*Summary, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		with l as (
			select endpoint_id, max(created_at) as latest from healthchecks group by endpoint_id
		)
		select
			ps.id,
			ps.domain,
			ps.success_code,
			coalesce(p.resp_status, 0) as resp_status,
			max(coalesce(p.created_at, '0001-01-01')) as last_check
		from endpoints ps
		left outer join l on l.endpoint_id = ps.id
		left outer join healthchecks p on
			p.endpoint_id = ps.id
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

// GetAll returns all Endpoints for all users.
// This is used internally. It is not exposed via the API as it would required admin privileges.
func (pg *PostgresRepo) GetAll(ctx context.Context) ([]*Endpoint, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `select id, domain, success_code from endpoints`
	rows, err := pg.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := []*Endpoint{}
	for rows.Next() {
		p := &Endpoint{}
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

// GetByID returns an Endpoint by its ID.
func (pg *PostgresRepo) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*Endpoint, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := "select id, domain, success_code from endpoints where id = $1 and user_id = $2"
	row := pg.DB.QueryRow(ctx, q, id, userID)

	p := &Endpoint{}

	err := row.Scan(&p.ID, &p.Domain, &p.SuccessCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, validation.ErrNotFound
		}
		return nil, err
	}

	return p, nil
}

// GetHealthcheckByID returns all healthchecks for an Endpoint.
func (pg *PostgresRepo) GetHealthcheckByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) ([]*Healthcheck, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
	select p.id, p.endpoint_id, p.resp_status, p.took_ms, p.created_at
		from healthchecks p
		inner join endpoints ps
			on ps.id = p.endpoint_id
		where p.endpoint_id = $1
			and ps.user_id = $2
		order by p.created_at desc;
	`

	rows, err := pg.DB.Query(ctx, q, id, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checks := []*Healthcheck{}

	for rows.Next() {
		p := &Healthcheck{}
		err := rows.Scan(&p.ID, &p.EndpointID, &p.RespStatus, &p.TookMs, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		checks = append(checks, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return checks, nil
}

// Count returns the number of Endpoints for a user.
func (pg *PostgresRepo) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var count int
	err := pg.DB.QueryRow(ctx, "select count(*) from endpoints where user_id = $1", userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// DeleteByID deletes an Endpoint and all its Healthchecks.
func (pg *PostgresRepo) DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := pg.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	resp, err := tx.Exec(ctx, "delete from endpoints where id = $1 and user_id = $2", id, userID)
	if err != nil || resp.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "delete from healthchecks where endpoint_id = $1", id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

// DeleteOldHealthchecks deletes all Healthchecks older than the given age.
func (pg *PostgresRepo) DeleteOldHealthchecks(ctx context.Context, age time.Duration) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := "delete from healthchecks where created_at < $1"
	arg := time.Now().UTC().Add(-age).Format(time.DateTime)
	resp, err := pg.DB.Exec(ctx, q, arg)
	if err != nil {
		return 0, err
	}
	return resp.RowsAffected(), nil
}
