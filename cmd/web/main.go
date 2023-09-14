// Package contains the web server.
package main

import (
	"domainator/internal/certs"
	"domainator/internal/config"
	"domainator/internal/db"
	"domainator/internal/endpoints"
	"domainator/internal/events"
	"domainator/internal/logger"
	"domainator/internal/plans"
	"domainator/internal/users"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

func init() {
	config.LoadEnv()
	logger.Init(os.Stdout, os.Stderr)
}

func main() {
	validate := validator.New()
	addr := fmt.Sprintf(":%d", config.GetInt("PORT"))
	srv, mux := buildServer(addr)
	db := db.MustInit(config.GetString("DSN"))
	defer db.Close()

	// Plans
	plansRepo := plans.NewPostgresRepo(db)

	// Users
	usersRepo := users.NewPostgresRepo(db)
	usersController := users.NewController(usersRepo, validate, plansRepo)
	users.AttachRoutes(mux, usersController)

	// Endpoints
	endpointsRepo := endpoints.NewPostgresRepo(db)
	endpointsController := endpoints.NewController(endpointsRepo, validate, plansRepo)
	endpoints.AttachRoutes(mux, endpointsController)

	// Certs
	certsRepo := certs.NewPostgresRepo(db)
	certsController := certs.NewController(certsRepo, validate, plansRepo)
	certs.AttachRoutes(mux, certsController)

	// Events
	eventsRepo := events.NewPostgresRepo(db)
	eventsController := events.NewController(eventsRepo, validate)
	events.AttachRoutes(mux, eventsController)

	// Start server
	logger.Writer.Info(fmt.Sprintf("Starting server on %s", addr))
	err := srv.ListenAndServe()
	logger.Writer.Fatal(err)
}
