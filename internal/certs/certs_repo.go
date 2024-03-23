package certs

import (
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

const QueryTimeout = 5 * time.Second

type Repo interface {
	Save(ctx context.Context, cert repoCert) error
	GetAll(ctx context.Context, userID ID) ([]repoCert, error)
	Get(ctx context.Context, id ID) (repoCert, error)
	Update(ctx context.Context, id ID, expiry time.Time, issuer string, updatedAt time.Time, e string) error
	Delete(ctx context.Context, id ID) error
}

type CertsRepo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *CertsRepo {
	return &CertsRepo{db}
}

func (r *CertsRepo) Save(ctx context.Context, cert repoCert) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `insert into certificates (id, user_id, domain, issuer, expires_at)
    values ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, q, cert.ID, cert.UserID, cert.Domain, cert.Issuer, cert.ExpiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateDomain
		}
		return err
	}

	return nil
}

func (r *CertsRepo) GetAll(ctx context.Context, userID ID) ([]repoCert, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `
    select
      id, user_id, domain, issuer, expires_at, created_at, updated_at, coalesce(error, '') as error
    from
      certificates
    where
      user_id = $1`

	rows, _ := r.db.Query(ctx, q, userID)
	return pgx.CollectRows(rows, pgx.RowToStructByName[repoCert])
}

func (r *CertsRepo) Get(ctx context.Context, id ID) (repoCert, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `
    select
      id, user_id, domain, issuer, expires_at, created_at, updated_at, coalesce(error, '') as error
    from
      certificates
    where
      id = $1`

	rows, _ := r.db.Query(ctx, q, id)
	cert, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[repoCert])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repoCert{}, ErrNotFound
		}
		return repoCert{}, err
	}

	return cert, nil
}

func (r *CertsRepo) Update(ctx context.Context, id ID, expiry time.Time, issuer string, updatedAt time.Time, e string) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `
    update
      certificates
    set
      issuer = $2,
      expires_at = $3,
      updated_at = $4,
      error = nullif($5, '')
    where
      id = $1`

	res, err := r.db.Exec(ctx, q, id, issuer, expiry, updatedAt, e)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// TODO: make it a soft delete
func (r *CertsRepo) Delete(ctx context.Context, id ID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `delete from certificates where id = $1`
	res, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
