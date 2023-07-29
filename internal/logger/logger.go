// Package logger provides a simple logger interface.
package logger

import (
	"fmt"
	"io"
	"log"
	"strings"
)

// Logit is a simple logger interface.
type Logit struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// Writer is a global logger instance that any package can use directly.
var Writer *Logit

// Init creates a logger instance and assigns it to the global variable `Writer`.
func Init(infoOut io.Writer, errOut io.Writer) {
	Writer = &Logit{
		InfoLog:  log.New(infoOut, "INFO\t", log.LUTC|log.Ltime|log.Lshortfile),
		ErrorLog: log.New(errOut, "ERROR\t", log.LUTC|log.Ltime|log.Lshortfile),
	}
}

// Info logs an info message.
func (l *Logit) Info(msgs ...any) {
	l.InfoLog.Output(2, join(msgs))
}

// Error logs an error message.
func (l *Logit) Error(msgs ...any) {
	l.ErrorLog.Output(2, join(msgs))
}

// Fatal logs an error message and exits.
func (l *Logit) Fatal(msgs ...any) {
	l.ErrorLog.Output(2, join(msgs))
}

func join(values []any) string {
	var msg string
	for _, v := range values {
		msg += fmt.Sprintf("%v ", v)
	}
	return strings.Trim(msg, " ")
}
