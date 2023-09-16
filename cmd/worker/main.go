// Package contains workers to perform background tasks.
package main

import (
	"domainator/internal/config"
	"domainator/internal/db"
	"domainator/internal/inspector"
	"domainator/internal/logger"
	"os"
)

func init() {
	config.LoadEnv()
	logger.Init(os.Stdout, os.Stderr)
}

func main() {
	logger.Writer.Info("Worker started")

	db := db.MustInit(config.GetString("DSN"))
	logger.Writer.Info("DB connection established")

	worker := inspector.New(db)
	worker.Start()
	logger.Writer.Info("Worker ended")

	db.Close()
	logger.Writer.Info("DB connection closed")
}
