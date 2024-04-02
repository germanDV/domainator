package common

import (
	"errors"
	"log/slog"
	"os"
)

func GetLogger(format string) (*slog.Logger, error) {
	switch format {
	case "text":
		return slog.New(slog.NewTextHandler(os.Stdout, nil)), nil
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stdout, nil)), nil
	default:
		return nil, errors.New("invalid log format, use one of 'text' or 'json'")
	}
}
