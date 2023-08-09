// Package db contains the database connection initialization.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Init establishes a PostgreSQL connection pool.
func Init(dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MaxConnIdleTime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

// MustInit calls Init and panics if there is an error.
func MustInit(dsn string) *pgxpool.Pool {
	pool, err := Init(dsn)
	if err != nil {
		panic(err)
	}
	return pool
}
