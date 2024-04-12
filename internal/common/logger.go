package common

import (
	"errors"
	"log/slog"
	"os"
)

var levels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func GetLogger(format string, level string) (*slog.Logger, error) {
	lvl, ok := levels[level]
	if !ok {
		return nil, errors.New("invalid log level, use one of 'debug', 'info', 'warn', 'error'")
	}

	opts := &slog.HandlerOptions{Level: lvl}
	switch format {
	case "text":
		return slog.New(slog.NewTextHandler(os.Stdout, opts)), nil
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, opts)), nil
	default:
		return nil, errors.New("invalid log format, use one of 'text' or 'json'")
	}
}
