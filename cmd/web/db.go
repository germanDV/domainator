package main

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func openDB(dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 15 * time.Minute

	// pool, err := pgxpool.New(context.Background(), dsn)
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
