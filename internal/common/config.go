package common

import (
	"os"

	"github.com/germandv/domainator/internal/configstruct"
)

func GetConfig[T any]() (*T, error) {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	config := new(T)
	if env != "prod" {
		err := configstruct.LoadAndParse(config, "./.env")
		if err != nil {
			return nil, err
		}
	} else {
		err := configstruct.Parse(config)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}
