// Package bg provides utilities to work with goroutines.
package bg

import (
	"domainator/internal/logger"
	"fmt"
	"runtime/debug"
)

// Run runs the task in a goroutine and recovers from panics.
func Run(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				trace := fmt.Sprintf("%v\n%s", err, debug.Stack())
				logger.Writer.Error(trace)
			}
		}()

		fn()
	}()
}
