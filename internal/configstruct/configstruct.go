package configstruct

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	errEnvFileNotFound = errors.New("env file not found")
	errNoGoModFile     = errors.New("could not find go.mod file")
)

// Parse takes a pointer to a struct and uses the 'env' and 'default'
// struct tags to populate it with values from environment variables.
//
// Before reading from environment variables, it will load the env file
// if it exists. The path to the env file is relative to the project root.
// To determine the project root, it will walk back from the current directory
// until it finds a go.mod file. If no go.mod file is found, it will ignore this step
// and continue reading from environment variables.
//
// Variables are considered required and will return an error unless
// a default value is provided.
//
// Supported types are: string, int, bool and time.Duration.
// Nested structs are not supported.
func Parse[T any](configStruct *T, configFilepath string) error {
	err := fileToEnv(configFilepath)
	// "env file not found" and "no go.mod" are ignored as these are common cases on cloud.
	if err != nil && !errors.Is(err, errEnvFileNotFound) && errors.Is(err, errNoGoModFile) {
		return err
	}

	return envToStruct(configStruct)
}

// fileToEnv looks for the env file, parses it and loads variables into the environment.
func fileToEnv(configFilepath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	root, err := getRootPath(cwd)
	if err != nil {
		return err
	}

	path := filepath.Join(root, configFilepath)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errEnvFileNotFound
		}
		return err
	}
	defer f.Close()

	multiline := false
	multilineKey := ""
	multilineVal := ""

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if multiline {
			multilineVal += "\n" + line
			if strings.HasSuffix(multilineVal, `"`) {
				os.Setenv(multilineKey, strings.TrimSuffix(multilineVal, `"`))
				multilineKey = ""
				multilineVal = ""
				multiline = false
			}
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid line format: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if strings.HasPrefix(val, `"`) && !strings.HasSuffix(val, `"`) {
			multiline = true
			multilineKey = key
			multilineVal = strings.TrimPrefix(val, `"`)
		}

		val = strings.Trim(val, `"`)
		val = strings.Trim(val, `'`)

		_, exists := os.LookupEnv(key)
		if !exists {
			os.Setenv(key, val)
		}
	}

	return nil
}

// envToStruct gets values from envirnoment variables and sets them into the provided configStruct.
func envToStruct[T any](configStruct *T) error {
	v := reflect.TypeOf(*configStruct)

	for i := 0; i < v.NumField(); i++ {
		structField := v.Field(i).Name
		structFieldType := v.Field(i).Type
		envVarName := v.Field(i).Tag.Get("env")
		defaultValue := v.Field(i).Tag.Get("default")

		envVarValue, ok := os.LookupEnv(envVarName)
		if !ok {
			if defaultValue == "" {
				return fmt.Errorf("missing env var %v (no default provided)", envVarName)
			}
			envVarValue = defaultValue
		}

		value, err := cast(structFieldType.Name(), envVarValue)
		if err != nil {
			return err
		}

		reflect.ValueOf(configStruct).Elem().FieldByName(structField).Set(value)
	}

	return nil
}

// getRootPath returns the root path of the project,
// walking back from the current directory until it finds a go.mod file.
func getRootPath(path string) (string, error) {
	if path == "/" {
		return "", errNoGoModFile
	}
	_, err := os.Stat(filepath.Join(path, "go.mod"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	} else if err == nil {
		return path, nil
	} else {
		return getRootPath(filepath.Join(path, ".."))
	}
}

// cast takes a string and trys to cast the value to its intended type.
func cast(fieldType string, fieldValue string) (reflect.Value, error) {
	switch fieldType {
	case "string":
		return reflect.ValueOf(fieldValue), nil
	case "int":
		v, err := strconv.Atoi(fieldValue)
		if err != nil {
			e := fmt.Errorf("cannot parse %s as int: %w", fieldValue, err)
			return reflect.ValueOf(nil), e
		}
		return reflect.ValueOf(v), nil
	case "bool":
		v, err := strconv.ParseBool(fieldValue)
		if err != nil {
			e := fmt.Errorf("cannot parse %s as bool: %w", fieldValue, err)
			return reflect.ValueOf(nil), e
		}
		return reflect.ValueOf(v), nil
	case "Duration":
		v, err := time.ParseDuration(fieldValue)
		if err != nil {
			e := fmt.Errorf("cannot parse %s as time.Duration: %w", fieldValue, err)
			return reflect.ValueOf(nil), e
		}
		return reflect.ValueOf(v), nil
	default:
		return reflect.ValueOf(nil), fmt.Errorf("unsupported type %s", fieldType)
	}
}
