package main

import (
	"domainator/internal/db"
	"domainator/internal/services"
	"flag"
	"time"
)

var pinger services.Pinger

func main() {
	dsn := flag.String("dsn", "postgres://postgres:pass123@localhost:5432/domainator", "PostgreSQL data source name")
	flag.Parse()

	db := db.MustInit(*dsn)
	defer db.Close()
	pinger = &services.PingService{DB: db}

	startLoop()
}

func startLoop() {
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	defer close(quit)

	doPings()

	for {
		select {
		case <-ticker.C:
			doPings()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
