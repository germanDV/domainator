// Package logger provides a simple logger interface
package logger

import (
	"log"
	"os"
)

// Logit is a simple logger interface
type Logit struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// New returns a new Logit instance
func New() *Logit {
	return &Logit{
		InfoLog:  log.New(os.Stdout, "INFO\t", log.LUTC|log.Ltime|log.Lshortfile),
		ErrorLog: log.New(os.Stderr, "ERROR\t", log.LUTC|log.Ltime|log.Lshortfile),
	}
}

// Info logs an info message
func (l *Logit) Info(msgs ...any) {
	l.InfoLog.Println(msgs...)
}

// Error logs an error message
func (l *Logit) Error(msgs ...any) {
	l.ErrorLog.Println(msgs...)
}

// Fatal logs an error message and exits
func (l *Logit) Fatal(msgs ...any) {
	l.ErrorLog.Fatal(msgs...)
}
