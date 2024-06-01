package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/germandv/domainator/internal/cache"
	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/db"
	"github.com/germandv/domainator/internal/notifier"
	"github.com/germandv/domainator/internal/tlser"
	"github.com/germandv/domainator/internal/users"
)

const CacheKey = "domainator_worker_running"

type WorkerConfig struct {
	Env             string `env:"APP_ENV" default:"dev"`
	LogFormat       string `env:"LOG_FORMAT" default:"text"`
	LogLevel        string `env:"LOG_LEVEL" default:"info"`
	PostgresConnStr string `env:"POSTGRES_CONN_STR"`
	BatchSize       int    `env:"BATCH_SIZE" default:"50"`
	RedisHost       string `env:"REDIS_HOST" default:"localhost"`
	RedisPort       int    `env:"REDIS_PORT" default:"6379"`
	RedisPassword   string `env:"REDIS_PASSWORD" default:" "`
}

// This worker is meant to be run as a cron job,
// it will check all the certificates in the database and update their details,
// sending notifications for those that have expired or will expire soon.
func main() {
	config, err := common.GetConfig[WorkerConfig]()
	if err != nil {
		panic(err)
	}

	logger, err := common.GetLogger(config.LogFormat, config.LogLevel)
	if err != nil {
		panic(err)
	}

	err = exclusiveRun(config, logger, run)
	if err != nil {
		panic(err)
	}
}

func exclusiveRun(
	config *WorkerConfig,
	logger *slog.Logger,
	fn func(config *WorkerConfig, logger *slog.Logger) error,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cacheClient := cache.New(config.RedisHost, config.RedisPort, config.RedisPassword)

	alreadyRunning, err := cacheClient.Get(ctx, CacheKey)
	if err != nil && !errors.Is(err, cache.ErrNoKey) {
		return err
	}

	if alreadyRunning == "true" {
		logger.Info("Another instance of the worker is already running, exiting")
		return nil
	}

	err = cacheClient.Set(ctx, CacheKey, "true", 10*time.Minute)
	if err != nil {
		return err
	}

	err = fn(config, logger)

	e := cacheClient.Set(context.Background(), CacheKey, "false", 0)
	if e != nil {
		logger.Warn("failed to set cache key to false", "error", e.Error())
	}

	return err
}

func run(config *WorkerConfig, logger *slog.Logger) error {
	logger.Info("Starting worker")

	db, err := db.InitWithConnStr(config.PostgresConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %s", err)
	}

	certsRepo := certs.NewRepo(db)
	tlsClient := tlser.New(5 * time.Second)
	certsService := certs.NewService(tlsClient, certsRepo, 10)

	usersRepo := users.NewRepo(db)
	usersService := users.NewService(usersRepo)

	slacker := notifier.NewSlacker()

	doneCh := make(chan struct{})
	errCh := make(chan error)
	notificationCh := make(chan notifier.Notification, 10)

	go func() {
		err = certsService.ProcessBatch(
			context.Background(),
			config.BatchSize,
			notificationCh,
			logger,
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
			return fmt.Errorf("failed to process batch: %s", err)
		case <-doneCh:
			logger.Info("Batch processed successfully")
			return nil
		case n := <-notificationCh:
			logger.Debug("Sending notification", "domain", n.Domain, "status", n.Status, "hours", n.Hours)
			userID, err := common.ParseID(n.UserID)
			if err != nil {
				logger.Error("Failed to parse user ID", "id", n.UserID, "error", err.Error())
				continue
			}

			user, err := usersService.GetByID(context.Background(), users.GetByIDReq{UserID: userID})
			if err != nil {
				logger.Error("Failed to fetch user data", "id", n.UserID, "error", err.Error())
				continue
			}

			if user.WebhookURL.String() == "" {
				logger.Debug("User has not provided a webhook, skipping notification", "id", n.UserID)
				continue
			}

			err = slacker.Notify(user.WebhookURL.String(), n)
			if err != nil {
				logger.Error("Failed to send notification", "id", n.UserID, "webhook_url", user.WebhookURL.String(), "error", err.Error())
			}
		}
	}
}
