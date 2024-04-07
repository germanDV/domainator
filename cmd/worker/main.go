package main

import (
	"context"
	"os"
	"time"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/db"
	"github.com/germandv/domainator/internal/notifier"
	"github.com/germandv/domainator/internal/tlser"
)

type WorkerConfig struct {
	Env              string `env:"APP_ENV" default:"dev"`
	LogFormat        string `env:"LOG_FORMAT"`
	PostgresHost     string `env:"POSTGRES_HOST"`
	PostgresPort     int    `env:"POSTGRES_PORT"`
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresDatabase string `env:"POSTGRES_DB"`
	BatchSize        int    `env:"BATCH_SIZE" default:"50"`
	Concurrency      int    `env:"BATCH_SIZE" default:"20"`
	SlackTestURL     string `env:"SLACK_TEST_WEBHOOK_URL" default:" "`
}

// This worker is meant to be run as a cron job,
// it will check all the certificates in the database and update their details,
// sending notifications for those that have expired or will expire soon.
func main() {
	config, err := common.GetConfig[WorkerConfig]()
	if err != nil {
		panic(err)
	}

	logger, err := common.GetLogger(config.LogFormat)
	if err != nil {
		panic(err)
	}

	logger.Info("Starting worker")

	db, err := db.Init(
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDatabase,
		config.Env != "dev",
	)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	certsRepo := certs.NewRepo(db)
	tlsClient := tlser.New(5 * time.Second)
	certsService := certs.NewService(tlsClient, certsRepo)
	slacker := notifier.NewSlacker()

	doneCh := make(chan struct{})
	errCh := make(chan error)
	notificationCh := make(chan notifier.Notification, 10)

	go func() {
		err = certsService.ProcessBatch(
			context.Background(),
			config.BatchSize,
			config.Concurrency,
			notificationCh,
		)
		if err != nil {
			errCh <- err
		} else {
			doneCh <- struct{}{}
		}
	}()

	for {
		select {
		case err := <-errCh:
			logger.Error("Failed to process batch", "error", err)
			os.Exit(1)
		case <-doneCh:
			logger.Info("Batch processed successfully")
			os.Exit(0)
		case n := <-notificationCh:
			logger.Debug("Sending notification", "domain", n.Domain, "status", n.Status, "hours", n.Hours)
			// TODO: allow users to enter a Slack Webhook URL
			slacker.Notify(config.SlackTestURL, n)
		}
	}
}
