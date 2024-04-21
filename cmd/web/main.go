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
	"github.com/germandv/domainator/internal/db"
	"github.com/germandv/domainator/internal/githubauth"
	"github.com/germandv/domainator/internal/handlers"
	"github.com/germandv/domainator/internal/tlser"
	"github.com/germandv/domainator/internal/tokenauth"
	"github.com/germandv/domainator/internal/users"
)

type AppConfig struct {
	Env             string `env:"APP_ENV" default:"dev"`
	LogFormat       string `env:"LOG_FORMAT"`
	LogLevel        string `env:"LOG_LEVEL" default:"info"`
	Port            int    `env:"PORT"`
	AuthPublKey     string `env:"AUTH_PUBLIC_KEY"`
	AuthPrivKey     string `env:"AUTH_PRIVATE_KEY"`
	RedisHost       string `env:"REDIS_HOST"`
	RedisPort       int    `env:"REDIS_PORT"`
	RedisPassword   string `env:"REDIS_PASSWORD" default:" "`
	PostgresConnStr string `env:"POSTGRES_CONN_STR"`
	GithubClientID  string `env:"GITHUB_CLIENT_ID"`
	GithubSecret    string `env:"GITHUB_SECRET"`
	Host            string `env:"HOST" default:"http://localhost"`
}

func main() {
	config, err := common.GetConfig[AppConfig]()
	if err != nil {
		panic(err)
	}

	logger, err := common.GetLogger(config.LogFormat, config.LogLevel)
	if err != nil {
		panic(err)
	}

	db, err := db.InitWithConnStr(config.PostgresConnStr)
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
	mux.Handle("GET /settings", authz(handlers.GetSettings(usersService)))
	mux.Handle("POST /settings/webhook", authz(handlers.SetWebhookURL(usersService)))

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
