package main

import (
	"fmt"
	"os"

	"github.com/germandv/domainator/internal/configstruct"
	"github.com/germandv/domainator/internal/tokenauth"
)

type AppConfig struct {
	AuthPublKey string `env:"AUTH_PUBLIC_KEY"`
	AuthPrivKey string `env:"AUTH_PRIVATE_KEY"`
}

// Generate auth token
func main() {
	if len(os.Args) < 2 {
		panic("provide a user ID")
	}
	userID := os.Args[1]

	config, err := getConfig()
	if err != nil {
		panic(err)
	}

	tokenService, err := tokenauth.New(config.AuthPrivKey, config.AuthPublKey)
	if err != nil {
		panic(err)
	}

	token, err := tokenService.Generate(userID, "")
	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}

func getConfig() (*AppConfig, error) {
	config := AppConfig{}
	err := configstruct.Parse(&config, "./.env")
	if err != nil {
		return nil, err
	}
	return &config, nil
}
