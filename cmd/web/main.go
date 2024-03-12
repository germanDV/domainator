package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/germandv/domainator/internal/cache"
	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/configstruct"
	"github.com/germandv/domainator/internal/handlers"
	"github.com/germandv/domainator/internal/middleware"
	"github.com/germandv/domainator/internal/tlser"
	"github.com/germandv/domainator/internal/tokenauth"
)

type AppConfig struct {
	LogFormat     string `env:"LOG_FORMAT"`
	Port          int    `env:"PORT"`
	AuthPublKey   string `env:"AUTH_PUBLIC_KEY"`
	AuthPrivKey   string `env:"AUTH_PRIVATE_KEY"`
	RedisHost     string `env:"REDIS_HOST"`
	RedisPort     int    `env:"REDIS_PORT"`
	RedisPassword string `env:"REDIS_PASSWORD" default:" "`
}

func main() {
	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	logger, err := getLogger(config.LogFormat)
	if err != nil {
		panic(err)
	}

	cacheClient := cache.New(config.RedisHost, config.RedisPort, config.RedisPassword)
	tlsClient := tlser.New(5 * time.Second)
	certsService := certs.NewService(tlsClient)
	fileServer := http.FileServer(http.Dir("./static"))

	authService, err := tokenauth.New(config.AuthPrivKey, config.AuthPublKey)
	if err != nil {
		panic(err)
	}
	authn := middleware.AuthBuilder(authService, false)
	authz := middleware.AuthBuilder(authService, true)

	mux := http.NewServeMux()
	mux.Handle("GET /static/*", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("GET /healthcheck", handlers.GetHealthcheck(cacheClient))
	mux.Handle("GET /", authn(handlers.GetHome(certsService)))
	mux.HandleFunc("GET /login", handlers.GetAccess())
	mux.Handle("POST /domain", authz(handlers.RegisterDomain(certsService)))
	mux.Handle("PUT /domain/{id}", authz(handlers.UpdateDomain(certsService)))
	mux.Handle("DELETE /domain/{id}", authz(handlers.DeleteDomain(certsService)))

	addr := fmt.Sprintf(":%d", config.Port)
	commonMiddleware := middleware.CommonBuilder(logger, cacheClient)
	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		Handler:      commonMiddleware(mux),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	killSig := make(chan os.Signal, 1)
	signal.Notify(killSig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Info("Server closed")
			} else {
				logger.Error("Server error", "err", err)
				os.Exit(1)
			}
		}
	}()

	logger.Info("Starting server", "addr", addr)

	<-killSig

	logger.Info("Shutting down server")

	err = cacheClient.Close()
	if err != nil {
		logger.Error("Error closing redis", "err", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	if err != nil {
		logger.Error("Failed to shut down gracefully", "err", err)
		os.Exit(1)
	}

	logger.Info("Server shutdown complete")
}

func getConfig() (*AppConfig, error) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	config := AppConfig{}
	if env != "prod" {
		err := configstruct.LoadAndParse(&config, "./.env")
		if err != nil {
			return nil, err
		}
	} else {
		err := configstruct.Parse(&config)
		if err != nil {
			return nil, err
		}
	}

	return &config, nil
}

func getLogger(format string) (*slog.Logger, error) {
	switch format {
	case "text":
		return slog.New(slog.NewTextHandler(os.Stdout, nil)), nil
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, nil)), nil
	default:
		return nil, errors.New("invalid log format, use one of 'text' or 'json'")
	}
}
