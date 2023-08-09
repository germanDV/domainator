package certs

import (
	"context"
	"domainator/internal/certstatus"
	"domainator/internal/config"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo is an interface that represents a repository for TLS certificates.
type Repo interface {
	Save(ctx context.Context, userID uuid.UUID, payload *CreateCertReq) (uuid.UUID, error)
	SaveCheck(ctx context.Context, check *Check) error
	GetSummary(ctx context.Context, userID uuid.UUID) ([]*Summary, error)
	GetAll(ctx context.Context) ([]*Cert, error)
	Count(ctx context.Context, userID uuid.UUID) (int, error)
	DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	// TODO:
	// DeleteOldChecks(ctx context.Context, age time.Duration) (int64, error)
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

// Save persists a `Cert`, aka: a new domain whose TLS certification will be checked.
func (pg *PostgresRepo) Save(ctx context.Context, userID uuid.UUID, payload *CreateCertReq) (uuid.UUID, error) {
	newID := uuid.New()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into certs (id, user_id, domain) values ($1, $2, $3)`
	_, err := pg.DB.Exec(ctx, q, newID, userID, payload.Domain)
	if err != nil {
		return uuid.Nil, err
	}

	return newID, nil
}

// SaveCheck persists a check result for a certificate.
func (pg *PostgresRepo) SaveCheck(ctx context.Context, c *Check) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into certchecks (id, cert_id, resp_status, expiry) values ($1, $2, $3, $4)`
	_, err := pg.DB.Exec(ctx, q, c.ID, c.CertID, c.RespStatus.String(), c.Expiry)
	if err != nil {
		return err
	}

	return nil
}

// GetSummary returns all Certs for a user with their latest checks.
func (pg *PostgresRepo) GetSummary(ctx context.Context, userID uuid.UUID) ([]*Summary, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `
		with l as (
			select cert_id, max(created_at) as latest from certchecks group by cert_id
		)
		select
			c.id,
			c.domain,
			coalesce(cc.resp_status, '') as resp_status,
			max(coalesce(cc.expiry, '0001-01-01')) as expiry,
			max(coalesce(cc.created_at, '0001-01-01')) as last_check
		from certs c
		left outer join l on l.cert_id = c.id
		left outer join certchecks cc on cc.cert_id = c.id and cc.created_at = l.latest
		where c.user_id = $1
		group by c.id, cc.resp_status;
	`

	rows, err := pg.DB.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	threshold := time.Now().Add(config.GetDuration("CERT_EXPIRY_THRESHOLD"))

	summaries := []*Summary{}
	for rows.Next() {
		s := &Summary{}
		var respStatus string

		err := rows.Scan(&s.ID, &s.Domain, &respStatus, &s.Expiry, &s.LastCheck)
		if err != nil {
			return nil, err
		}

		switch respStatus {
		case certstatus.CannotConnect.String():
			s.Status = certstatus.CannotConnect
		case certstatus.HostnameMismatch.String():
			s.Status = certstatus.HostnameMismatch
		case certstatus.OK.String():
			if s.Expiry.Before(time.Now()) {
				s.Status = certstatus.Expired
			} else if s.Expiry.Before(threshold) {
				s.Status = certstatus.AboutToExpire
			} else {
				s.Status = certstatus.OK
			}
		default:
			s.Status = certstatus.Nil
		}

		summaries = append(summaries, s)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return summaries, nil
}

// GetAll returns all Certs in the database.
// This is used internally. It is not exposed via the API as it would required admin privileges.
func (pg *PostgresRepo) GetAll(ctx context.Context) ([]*Cert, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	rows, err := pg.DB.Query(ctx, "select id, user_id, domain from certs")
	if err != nil {
		return nil, err
	}

	certs := []*Cert{}
	for rows.Next() {
		c := &Cert{}
		err := rows.Scan(&c.ID, &c.UserID, &c.Domain)
		if err != nil {
			return nil, err
		}
		certs = append(certs, c)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return certs, nil
}

// Count returns the number of Certs for a user.
func (pg *PostgresRepo) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var count int
	err := pg.DB.QueryRow(ctx, "select count(*) from certs where user_id = $1", userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// DeleteByID deletes a Cert and all of its checks.
func (pg *PostgresRepo) DeleteByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := pg.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	resp, err := tx.Exec(ctx, "delete from certs where id = $1 and user_id = $2", id, userID)
	if err != nil || resp.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "delete from certchecks where cert_id = $1", id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}
