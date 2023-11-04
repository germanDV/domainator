package plans

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo is an interface that the plans repository must implement.
type Repo interface {
	GetAll(ctx context.Context) ([]*Plan, error)
	GetByID(ctx context.Context, id int) (*Plan, error)
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

// GetAll returns all plans.
func (pg *PostgresRepo) GetAll(ctx context.Context) ([]*Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := "select id, name, price_cents, domain_limit, certs_limit from plans"
	rows, err := pg.DB.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	plans := []*Plan{}
	for rows.Next() {
		p := &Plan{}
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.DomainsLimit, &p.CertsLimit)
		if err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return plans, nil
}

// Getter is a function that returns a plan by its ID.
type Getter func(ctx context.Context, id int) (*Plan, error)

// GetByID returns a plan by its ID.
func (pg *PostgresRepo) GetByID(ctx context.Context, id int) (*Plan, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	p := Plan{}
	q := "select id, name, price_cents, domain_limit, certs_limit from plans where id = $1"
	err := pg.DB.QueryRow(ctx, q, id).Scan(&p.ID, &p.Name, &p.Price, &p.DomainsLimit, &p.CertsLimit)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
