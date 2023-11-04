package main

import (
	"domainator/internal/bg"
	"domainator/internal/certs"
	"domainator/internal/config"
	"domainator/internal/db"
	"domainator/internal/endpoints"
	"domainator/internal/events"
	"domainator/internal/plans"
	"domainator/internal/users"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
)

func init() {
	config.LoadEnv()
}

func main() {
	var logHandler slog.Handler
	switch config.GetString("LOG_FORMAT") {
	case "text":
		logHandler = slog.NewTextHandler(os.Stdout, nil)
	case "json":
		logHandler = slog.NewJSONHandler(os.Stdout, nil)
	default:
		panic("invalid log format, use one of 'text' or 'json'")
	}
	logger := slog.New(logHandler)

	bg.Init(logger)
	validate := validator.New()

	addr := fmt.Sprintf(":%d", config.GetInt("PORT"))
	srv, mux := buildServer(addr, logger)

	db := db.MustInit(config.GetString("DSN"))
	defer db.Close()

	// Plans
	plansRepo := plans.NewPostgresRepo(db)

	// Users
	usersRepo := users.NewPostgresRepo(db)
	usersController := users.NewController(usersRepo, validate, plansRepo.GetByID, logger)
	users.AttachRoutes(mux, usersController)

	// Endpoints
	endpointsRepo := endpoints.NewPostgresRepo(db)
	endpointsController := endpoints.NewController(endpointsRepo, validate, plansRepo.GetByID, logger)
	endpoints.AttachRoutes(mux, endpointsController)

	// Certs
	certsRepo := certs.NewPostgresRepo(db)
	certsController := certs.NewController(certsRepo, validate, plansRepo.GetByID, logger)
	certs.AttachRoutes(mux, certsController)

	// Events
	eventsRepo := events.NewPostgresRepo(db)
	eventsController := events.NewController(eventsRepo, validate, logger)
	events.AttachRoutes(mux, eventsController)

	// Start server
	logger.Info("Starting server", "addr", addr)
	err := srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}
