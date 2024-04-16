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

// Parse takes a pointer to a struct and uses the 'env' and 'default'
// struct tags to populate it with values from environment variables.
//
// Variables are considered required and will return an error unless
// a default value is provided.
//
// Supported types are: string, int, bool and time.Duration.
// Nested structs are not supported.
func Parse[T any](configStruct *T) error {
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

// LoadAndParse loads env vars from an env file and calls `Parse` to populate the config struct.
// Provide file path relative to the root of the project (aka: wherever go.mod is located).
func LoadAndParse[T any](configStruct *T, configFilepath string) error {
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
		return err
	}
	defer f.Close()

	err = setFromFile(f)
	if err != nil {
		return err
	}

	return Parse(configStruct)
}

// getRootPath returns the root path of the project,
// walking back from the current directory until it finds a go.mod file.
func getRootPath(path string) (string, error) {
	if path == "/" {
		return "", errors.New("could not find go.mod file")
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

// setFromFile parses env file and sets values to the environment, without overwriting existing ones.
func setFromFile(f *os.File) error {
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
