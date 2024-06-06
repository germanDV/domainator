package certs

import (
	"errors"
	"time"

	"github.com/germandv/domainator/internal/common"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

const QueryTimeout = 5 * time.Second

type Repo interface {
	Save(ctx context.Context, cert repoCert) error
	GetAll(ctx context.Context, userID common.ID) ([]repoCert, error)
	GetBatch(ctx context.Context, size int, cursor string) ([]repoCert, error)
	Get(ctx context.Context, id common.ID) (repoCert, error)
	Count(ctx context.Context, userID common.ID, limit int) (int, error)
	Update(ctx context.Context, userID common.ID, id common.ID, expiry time.Time, issuer string, updatedAt time.Time) error
	UpdateWithError(ctx context.Context, userID common.ID, id common.ID, error string, updatedAt time.Time) error
	Delete(ctx context.Context, userID common.ID, id common.ID) error
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

func (r *CertsRepo) GetAll(ctx context.Context, userID common.ID) ([]repoCert, error) {
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

func (r *CertsRepo) Get(ctx context.Context, id common.ID) (repoCert, error) {
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

func (r *CertsRepo) update(ctx context.Context, query string, args ...any) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	res, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *CertsRepo) Update(
	ctx context.Context,
	userID common.ID,
	id common.ID,
	expiry time.Time,
	issuer string,
	updatedAt time.Time,
) error {
	q := `
    update
      certificates
    set
      issuer = $2,
      expires_at = $3,
      updated_at = $4,
      error = ''
    where
      id = $1 and user_id = $5`
	return r.update(ctx, q, id, issuer, expiry, updatedAt, userID)
}

func (r *CertsRepo) UpdateWithError(
	ctx context.Context,
	userID common.ID,
	id common.ID,
	error string,
	updatedAt time.Time,
) error {
	q := `
    update
      certificates
    set
      error = $3,
      updated_at = $4
    where
      id = $1 and user_id = $2`
	return r.update(ctx, q, id, userID, error, updatedAt)
}

func (r *CertsRepo) Delete(ctx context.Context, userID common.ID, id common.ID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	return pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		q := `insert into certificates_deleted (id, user_id, domain, issuer, error, expires_at, created_at, updated_at)
      select * from certificates where id = $1 and user_id = $2`
		res, err := tx.Exec(ctx, q, id, userID)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return ErrNotFound
		}

		q = `delete from certificates where id = $1 and user_id = $2`
		res, err = r.db.Exec(ctx, q, id, userID)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return ErrNotFound
		}

		return nil
	})
}

func (r *CertsRepo) GetBatch(ctx context.Context, size int, lastID string) ([]repoCert, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `
    select
      id, user_id, domain, issuer, expires_at, created_at, updated_at, coalesce(error, '') as error
    from certificates
    where id < $2
    order by id desc
    limit $1`

	if lastID == "" {
		q = `
      select
        id, user_id, domain, issuer, expires_at, created_at, updated_at, coalesce(error, '') as error
      from certificates
      order by id desc
      limit $1`
	}

	var rows pgx.Rows
	if lastID == "" {
		rows, _ = r.db.Query(ctx, q, size)
	} else {
		rows, _ = r.db.Query(ctx, q, size, lastID)
	}

	certs, err := pgx.CollectRows(rows, pgx.RowToStructByName[repoCert])
	if err != nil {
		return nil, err
	}

	return certs, nil
}

func (r *CertsRepo) Count(ctx context.Context, userID common.ID, limit int) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	count := 0
	q := "select count(*) from certificates where user_id = $1 limit $2"

	err := r.db.QueryRow(ctx, q, userID, limit).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
