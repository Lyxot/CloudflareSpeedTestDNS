package conf

import (
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToUpper(snake)
}

// LoadEnvConfig loads configuration from environment variables and updates the config struct.
func LoadEnvConfig(config *Config) {
	loadStructFromEnv("CFSTD", reflect.ValueOf(config).Elem())
}

func loadStructFromEnv(prefix string, val reflect.Value) {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		envVarName := prefix + "_" + toSnakeCase(fieldType.Name)

		if field.Kind() == reflect.Struct {
			loadStructFromEnv(envVarName, field)
		} else {
			setFieldFromEnv(field, envVarName)
		}
	}
}

func setFieldFromEnv(field reflect.Value, envVarName string) {
	if !field.CanSet() {
		return
	}

	envValue := os.Getenv(envVarName)
	if envValue == "" {
		return
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(envValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intValue, err := strconv.ParseInt(envValue, 10, 64); err == nil {
			field.SetInt(intValue)
		}
	case reflect.Float32, reflect.Float64:
		if floatValue, err := strconv.ParseFloat(envValue, 64); err == nil {
			field.SetFloat(floatValue)
		}
	case reflect.Bool:
		if boolValue, err := strconv.ParseBool(envValue); err == nil {
			field.SetBool(boolValue)
		}
	}
}
