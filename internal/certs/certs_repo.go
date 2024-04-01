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
	Get(ctx context.Context, id common.ID) (repoCert, error)
	Update(ctx context.Context, userID common.ID, id common.ID, expiry time.Time, issuer string, updatedAt time.Time, e string) error
	UpdateWithTx(ctx context.Context, tx pgx.Tx, userID common.ID, id common.ID, expiry time.Time, issuer string, updatedAt time.Time, e string) error
	Delete(ctx context.Context, userID common.ID, id common.ID) error
	ProcessBatch(ctx context.Context, size int, cursor string) ([]repoCert, pgx.Tx, error)
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

func (r *CertsRepo) Update(
	ctx context.Context,
	userID common.ID,
	id common.ID,
	expiry time.Time,
	issuer string,
	updatedAt time.Time,
	e string,
) error {
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
      id = $1 and user_id = $6`

	res, err := r.db.Exec(ctx, q, id, issuer, expiry, updatedAt, e, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *CertsRepo) UpdateWithTx(
	ctx context.Context,
	tx pgx.Tx,
	userID common.ID,
	id common.ID,
	expiry time.Time,
	issuer string,
	updatedAt time.Time,
	e string,
) error {
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
      id = $1 and user_id = $6`

	res, err := tx.Exec(ctx, q, id, issuer, expiry, updatedAt, e, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// TODO: make it a soft delete
func (r *CertsRepo) Delete(ctx context.Context, userID common.ID, id common.ID) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `delete from certificates where id = $1 and user_id = $2`
	res, err := r.db.Exec(ctx, q, id, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *CertsRepo) ProcessBatch(ctx context.Context, size int, lastID string) ([]repoCert, pgx.Tx, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}

	q := `
    select
      id, user_id, domain, issuer, expires_at, created_at, updated_at, coalesce(error, '') as error
    from certificates
    where id < $2
    order by id desc
    limit $1
    for update skip locked`

	if lastID == "" {
		q = `
      select
        id, user_id, domain, issuer, expires_at, created_at, updated_at, coalesce(error, '') as error
      from certificates
      order by id desc
      limit $1
      for update skip locked`
	}

	var rows pgx.Rows
	if lastID == "" {
		rows, _ = tx.Query(ctx, q, size)
	} else {
		rows, _ = tx.Query(ctx, q, size, lastID)
	}

	certs, err := pgx.CollectRows(rows, pgx.RowToStructByName[repoCert])
	if err != nil {
		tx.Rollback(ctx)
		return nil, nil, err
	}

	return certs, tx, nil
}
