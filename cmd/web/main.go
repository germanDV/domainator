package main

import (
	"domainator/internal/config"
	"domainator/internal/db"
	"domainator/internal/inspector"
	"domainator/internal/services"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	pingSvc       services.Pinger
	templateCache map[string]*template.Template
	formDecoder   *form.Decoder
	validate      *validator.Validate
}

func init() {
	config.LoadEnv()
}

func main() {
	errorLog := log.New(os.Stderr, "ERROR\t", log.LUTC|log.Ltime|log.Lshortfile)
	infoLog := log.New(os.Stdout, "INFO\t", log.LUTC|log.Ltime)
	validate := validator.New()
	addr := fmt.Sprintf(":%d", config.GetInt("PORT"))
	db := db.MustInit(config.GetString("DSN"))
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	pinger := &services.PingService{
		Validator: validate,
		DB:        db,
	}

	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		pingSvc:       pinger,
		templateCache: templateCache,
		formDecoder:   form.NewDecoder(),
		validate:      validate,
	}

	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     app.errorLog,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	inspctr := inspector.New(db, pinger, errorLog, infoLog)
	inspctr.Start()

	infoLog.Printf("Starting server on %s", addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
