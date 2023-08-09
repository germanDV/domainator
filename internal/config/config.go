// Package config loads env vars from file and provides convenience functions
package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// ProjectName is the name of the project and it's useful to find the root path.
const ProjectName = "domainator"

// LoadEnv loads env vars and panics if there's an error.
// If ENV_FILENAME is set, it will load env vars from ProjectName/ENV_FILENAME.
// Otherwise, it will load env vars from .env.
func LoadEnv() {
	envFilename, ok := os.LookupEnv("ENV_FILENAME")
	var err error
	if ok {
		rootPath := GetRootPath()
		err = godotenv.Load(rootPath + "/" + envFilename)
	} else {
		err = godotenv.Load()
	}
	if err != nil {
		panic(err)
	}
}

// get returns the value of the env var with the given key,
// it panics if the env var is not set
func get(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		panic("Missing required environment variable " + key)
	}
	return v
}

// GetString returns the value of the env var as a string,
// it panics if env var is not set
func GetString(key string) string {
	return get(key)
}

// GetInt returns the value of the env var as an int
// it panics if env var is not set or if it cannot be converted to int
func GetInt(key string) int {
	str := get(key)
	v, err := strconv.Atoi(str)
	if err != nil {
		panic(err)
	}
	return v
}

// GetBool returns the value of the env var as a bool
// it panics if env var is not set or if it cannot be converted to bool
func GetBool(key string) bool {
	str := get(key)
	v, err := strconv.ParseBool(str)
	if err != nil {
		panic(err)
	}
	return v
}

// GetDuration returns the value of the env var as a time.Duration
// it panics if env var is not set or if it cannot be converted to time.Duration
func GetDuration(key string) time.Duration {
	str := get(key)
	v, err := time.ParseDuration(str)
	if err != nil {
		panic(err)
	}
	return v
}

// GetRootPath returns the root path of the project using the ProjectName
func GetRootPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile("^.*" + ProjectName)
	root := re.FindString(cwd)
	if root == "" {
		panic(fmt.Errorf("Could not find root path; is the ProjectName %q correct?", ProjectName))
	}

	return root
}
