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
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/configstruct"
	"github.com/germandv/domainator/internal/db"
	"github.com/germandv/domainator/internal/githubauth"
	"github.com/germandv/domainator/internal/handlers"
	"github.com/germandv/domainator/internal/tlser"
	"github.com/germandv/domainator/internal/tokenauth"
	"github.com/germandv/domainator/internal/users"
)

type AppConfig struct {
	Env              string `env:"APP_ENV" default:"dev"`
	LogFormat        string `env:"LOG_FORMAT"`
	Port             int    `env:"PORT"`
	AuthPublKey      string `env:"AUTH_PUBLIC_KEY"`
	AuthPrivKey      string `env:"AUTH_PRIVATE_KEY"`
	RedisHost        string `env:"REDIS_HOST"`
	RedisPort        int    `env:"REDIS_PORT"`
	RedisPassword    string `env:"REDIS_PASSWORD" default:" "`
	PostgresHost     string `env:"POSTGRES_HOST"`
	PostgresPort     int    `env:"POSTGRES_PORT"`
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresDatabase string `env:"POSTGRES_DB"`
	GithubClientID   string `env:"GITHUB_CLIENT_ID"`
	GithubSecret     string `env:"GITHUB_SECRET"`
	Host             string `env:"HOST" default:"http://localhost"`
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

	db, err := db.Init(
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDatabase,
		config.Env != "dev",
	)
	if err != nil {
		panic(err)
	}

	usersRepo := users.NewRepo(db)
	usersService := users.NewService(usersRepo)
	certsRepo := certs.NewRepo(db)
	cacheClient := cache.New(config.RedisHost, config.RedisPort, config.RedisPassword)
	tlsClient := tlser.New(5 * time.Second)
	certsService := certs.NewService(tlsClient, certsRepo)
	fileServer := http.FileServer(http.Dir("./static"))

	authService, err := tokenauth.New(config.AuthPrivKey, config.AuthPublKey)
	if err != nil {
		panic(err)
	}
	authn := handlers.AuthMdwBuilder(authService, false)
	authz := handlers.AuthMdwBuilder(authService, true)

	stateStr := common.GenerateRandomString(32)
	githubCfg := githubauth.NewGithubConfig(
		config.GithubClientID,
		config.GithubSecret,
		fmt.Sprintf("%s:%d/github/callback", config.Host, config.Port),
	)

	mux := http.NewServeMux()
	mux.Handle("GET /static/*", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("GET /healthcheck", handlers.GetHealthcheck(cacheClient, db))
	mux.Handle("GET /", authn(handlers.GetHome(certsService)))
	mux.Handle("GET /login", authn(handlers.GetAccess()))
	mux.Handle("GET /github/login", authn(handlers.GithubLogin(stateStr, githubCfg)))
	mux.HandleFunc("GET /github/callback", handlers.GithubCallback(stateStr, githubCfg, authService, usersService))
	mux.HandleFunc("POST /logout", handlers.Logout())
	mux.Handle("POST /domain", authz(handlers.RegisterDomain(certsService)))
	mux.Handle("PUT /domain/{id}", authz(handlers.UpdateDomain(certsService)))
	mux.Handle("DELETE /domain/{id}", authz(handlers.DeleteDomain(certsService)))

	addr := fmt.Sprintf(":%d", config.Port)
	commonMiddleware := handlers.CommonMdwBuilder(logger, cacheClient)
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

	// Gracefully shutdown
	<-killSig
	logger.Info("Shutting down server")
	err = cacheClient.Close()
	if err != nil {
		logger.Error("Error closing redis", "err", err)
	}
	db.Close()
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
