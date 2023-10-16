// Package bg provides utilities to work with goroutines.
package bg

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

var logger *slog.Logger

func Init(l *slog.Logger) {
	logger = l
}

// Run runs the task in a goroutine and recovers from panics.
func Run(fn func()) {
	if logger == nil {
		panic("bg package has not been initialized")
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				trace := fmt.Sprintf("%v\n%s", err, debug.Stack())
				logger.Error(trace)
			}
		}()

		fn()
	}()
}
