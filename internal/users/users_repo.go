package users

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/net/context"
)

const QueryTimeout = 5 * time.Second

type Repo interface {
	Save(ctx context.Context, user repoUser) error
	GetByEmail(ctx context.Context, email Email) (repoUser, error)
}

type UsersRepo struct {
	db *pgxpool.Pool
}

func NewRepo(db *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{db}
}

func (r *UsersRepo) Save(ctx context.Context, user repoUser) error {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `insert into users (id, name, email, identity_provider, identity_provider_id)
    values ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, q, user.ID, user.Name, user.Email, user.IdentityProvider, user.IdentityProviderID)

	return err
}

func (r *UsersRepo) GetByEmail(ctx context.Context, email Email) (repoUser, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()

	q := `
    select
      id, name, email, created_at, identity_provider, identity_provider_id
    from
      users
    where
      email = $1`

	rows, _ := r.db.Query(ctx, q, email.value)
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[repoUser])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repoUser{}, ErrNotFound
		}
		return repoUser{}, err
	}

	return user, nil
}
