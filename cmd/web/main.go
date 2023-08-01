// Package main is the entry point for the application.
package main

import (
	"domainator/internal/bg"
	"domainator/internal/config"
	"domainator/internal/db"
	"domainator/internal/inspector"
	"domainator/internal/logger"
	"domainator/internal/pings"
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

	// Pings
	pingsRepo := pings.NewPostgresRepo(db)
	pingsController := pings.NewController(pingsRepo, validate, plansRepo)
	pings.AttachRoutes(mux, pingsController)

	// Inspector (background tasks)
	inspctr := inspector.New(db)
	bg.Run(inspctr.Start)

	// Start server
	logger.Writer.Info(fmt.Sprintf("Starting server on %s", addr))
	err := srv.ListenAndServe()
	logger.Writer.Fatal(err)
}
