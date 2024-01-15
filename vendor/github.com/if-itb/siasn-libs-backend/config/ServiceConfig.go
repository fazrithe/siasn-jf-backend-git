package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// The ServiceConfig allows creators of a service to interact with environment variables easily.
// To create a ServiceConfig, you just need to supply a Prefix and ArraySeparator, and use the
// methods available in this class.
//
// To automatically parse configuration into a struct without having to use individual getters,
// see ParseTo.
type ServiceConfig struct {
	// The Prefix is added to all the config name that is supplied in getter functions
	// such as the GetString or through the use struct tags.
	Prefix string
	// The token to use to separate string in environment variables into array.
	// Used by getters such as GetStringArray.
	ArraySeparator string
}

func (sc ServiceConfig) getConfigName(name string) string {
	return sc.Prefix + "_" + name
}

func (sc ServiceConfig) GetString(name string) (string, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	if !exist {
		return configData, ErrConfigNotFound
	}
	return configData, nil
}

func (sc ServiceConfig) GetStringArray(name string) ([]string, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	configDataArray := strings.Split(configData, sc.ArraySeparator)
	if !exist {
		return configDataArray, ErrConfigNotFound
	}

	return configDataArray, nil
}

func (sc ServiceConfig) GetInt(name string) (int, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	if !exist {
		return 0, ErrConfigNotFound
	}
	return strconv.Atoi(configData)
}

func (sc ServiceConfig) GetBool(name string) (bool, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	if !exist {
		return false, ErrConfigNotFound
	}
	return strconv.ParseBool(configData)
}

func (sc ServiceConfig) GetFloat32(name string) (float32, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	if !exist {
		return 0, ErrConfigNotFound
	}
	number, err := strconv.ParseFloat(configData, 32)
	return float32(number), err
}

func (sc ServiceConfig) GetStringWithDefault(name string, defaultValue string) (string, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	if !exist {
		return defaultValue, nil
	}
	return configData, nil
}

func (sc ServiceConfig) GetStringArrayWithDefault(name string, defaultValue []string) ([]string, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	configDataArray := strings.Split(configData, sc.ArraySeparator)
	if !exist {
		return defaultValue, nil
	}

	return configDataArray, nil
}

func (sc ServiceConfig) GetIntWithDefault(name string, defaultValue int) (int, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	if !exist {
		return defaultValue, nil
	}
	return strconv.Atoi(configData)
}

func (sc ServiceConfig) GetBoolWithDefault(name string, defaultValue bool) (bool, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	if !exist {
		return defaultValue, nil
	}
	return strconv.ParseBool(configData)
}

func (sc ServiceConfig) GetFloat32WithDefault(name string, defaultValue float32) (float32, error) {
	configData, exist := os.LookupEnv(sc.getConfigName(name))
	if !exist {
		return defaultValue, nil
	}
	number, err := strconv.ParseFloat(configData, 32)
	return float32(number), err
}

// ParseTo accepts a pointer to a struct with fields already tagged with `config` tags.
// The `config` tag value indicates the name of the configuration to retrieve from. For example, a struct
// field of type int with `config:"PORT"` tag and ServiceConfig.Prefix set with "WEB", will have the value retrieved
// from an environment variable "WEB_PORT", and automatically parsed as integer.
//
// When the environment variable does not exists, the field is skipped. This way you can supply a prefilled struct that
// already have default values initialized. If the environment variable for the field does not exist (not configured
// by administrator of the service), then default value is used.
func (sc ServiceConfig) ParseTo(obj interface{}) error {
	assertPointer(obj)

	v := reflect.ValueOf(obj)
	realV := reflect.Indirect(v)
	t := realV.Type()

	for i := 0; i < realV.NumField(); i++ {
		tag, ok := t.Field(i).Tag.Lookup("config")
		if !ok {
			continue
		}

		switch realV.Field(i).Interface().(type) {
		case int:
			val, err := sc.GetInt(tag)
			if err != nil {
				if err == ErrConfigNotFound {
					continue
				}

				return sc.reformatParseError(tag, err)
			}

			realV.Field(i).Set(reflect.ValueOf(val))
		case string:
			val, err := sc.GetString(tag)
			if err != nil {
				if err == ErrConfigNotFound {
					continue
				}

				return sc.reformatParseError(tag, err)
			}

			realV.Field(i).Set(reflect.ValueOf(val))
		case float32:
			val, err := sc.GetFloat32(tag)
			if err != nil {
				if err == ErrConfigNotFound {
					continue
				}

				return sc.reformatParseError(tag, err)
			}

			realV.Field(i).Set(reflect.ValueOf(val))
		case bool:
			val, err := sc.GetBool(tag)
			if err != nil {
				if err == ErrConfigNotFound {
					continue
				}

				return sc.reformatParseError(tag, err)
			}

			realV.Field(i).Set(reflect.ValueOf(val))
		case []string:
			val, err := sc.GetStringArray(tag)
			if err != nil {
				if err == ErrConfigNotFound {
					continue
				}

				return sc.reformatParseError(tag, err)
			}

			realV.Field(i).Set(reflect.ValueOf(val))
		}
	}

	return nil
}

func (sc ServiceConfig) reformatParseError(name string, err error) error {
	return fmt.Errorf("cannot parse %s_%s: %v", sc.Prefix, name, err)
}

func assertPointer(value interface{}) {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic("given value is not a pointer, or nil")
	}
}
