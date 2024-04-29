package common

import (
	"github.com/germandv/domainator/internal/configstruct"
)

func GetConfig[T any]() (*T, error) {
	config := new(T)
	err := configstruct.Parse(config, ".env")
	if err != nil {
		return nil, err
	}

	return config, nil
}
