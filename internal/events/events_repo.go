package events

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo is an interface that represents a repository for events.
type Repo interface {
	Save(ctx context.Context, userID uuid.UUID, payload *CreateEventReq) (uuid.UUID, error)
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

// Save persists a new event.
func (pg *PostgresRepo) Save(ctx context.Context, userID uuid.UUID, ev *CreateEventReq) (uuid.UUID, error) {
	newID := uuid.New()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	q := `insert into events (id, user_id, name, payload) values ($1, $2, $3, $4)`
	_, err := pg.DB.Exec(ctx, q, newID, userID, ev.Name, ev.Payload)
	if err != nil {
		return uuid.Nil, err
	}

	return newID, nil
}
