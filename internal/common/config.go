package common

import (
	"os"

	"github.com/germandv/domainator/internal/configstruct"
)

func GetConfig[T any]() (*T, error) {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}

	config := new(T)

	err := configstruct.Parse(config, envFile)
	if err != nil {
		return nil, err
	}

	return config, nil
}
