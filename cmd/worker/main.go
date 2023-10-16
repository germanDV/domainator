// Package contains workers to perform background tasks.
package main

import (
	"domainator/internal/config"
	"domainator/internal/db"
	"domainator/internal/inspector"
	"log/slog"
	"os"
)

func init() {
	config.LoadEnv()
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Worker started")

	db := db.MustInit(config.GetString("DSN"))
	logger.Info("DB connection established")

	worker := inspector.New(db)
	worker.Start()
	logger.Info("Worker ended")

	db.Close()
	logger.Info("DB connection closed")
}
