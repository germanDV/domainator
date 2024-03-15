package certs

import (
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

const AnonymousUser = "00000000-0000-0000-0000-000000000000"

type Repo interface {
	Save(context.Context, Cert) error
	Get(id ID) (Cert, error)
	GetAll() ([]Cert, error)
	Delete(id ID) error
	Update(id ID, expiry time.Time, issuer Issuer, e string) (Cert, error)
}

type CertsRepo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *CertsRepo {
	return &CertsRepo{db}
}

func (r *CertsRepo) Save(ctx context.Context, cert Cert) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into certificates (id, user_id, domain, issuer, expires_at)
    values ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, q, cert.ID, AnonymousUser, cert.Domain, cert.Issuer, cert.ExpiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrDuplicateDomain
		}
		return err
	}

	return nil
}

func (r *CertsRepo) Get(id ID) (Cert, error) {
	return Cert{}, ErrNotFound
}

func (r *CertsRepo) GetAll() ([]Cert, error) {
	certs := make([]Cert, 0, 0)
	return certs, nil
}

func (r *CertsRepo) Delete(id ID) error {
	return nil
}

func (r *CertsRepo) Update(id ID, expiry time.Time, issuer Issuer, e string) (Cert, error) {
	return Cert{}, ErrNotFound
}
