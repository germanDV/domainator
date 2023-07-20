// Package main is the entry point for the application.
package main

import (
	"domainator/internal/config"
	"domainator/internal/db"
	"domainator/internal/inspector"
	"domainator/internal/logger"
	"domainator/internal/notifier"
	"domainator/internal/services"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

type application struct {
	logit         *logger.Logit
	pingSvc       services.Pinger
	userSvc       services.IUserService
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
	validate      *validator.Validate
	inspector     inspector.Inspector
	mailer        notifier.Notifier
}

func init() {
	config.LoadEnv()
}

func main() {
	validate := validator.New()
	addr := fmt.Sprintf(":%d", config.GetInt("PORT"))
	logit := logger.New()
	db := db.MustInit(config.GetString("DSN"))
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logit.Fatal(err)
	}

	pinger := &services.PingService{
		Validator: validate,
		DB:        db,
	}

	userSvc := &services.UserService{
		Validator: validate,
		DB:        db,
	}

	app := &application{
		logit:         logit,
		pingSvc:       pinger,
		userSvc:       userSvc,
		templateCache: templateCache,
		formDecoder:   form.NewDecoder(),
		validate:      validate,
		inspector:     inspector.New(db, pinger, logit),
		mailer:        notifier.NewMailer(),
	}

	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     app.logit.ErrorLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logit.Error(fmt.Sprintf("Inspector panicked! %v", err))
			}
		}()
		app.startInspector()
	}()

	logit.Info(fmt.Sprintf("Starting server on %s", addr))
	err = srv.ListenAndServe()
	logit.Fatal(err)
}
