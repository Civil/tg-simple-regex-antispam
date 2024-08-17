package config

import (
	"fmt"
)

func GetOptionString(config map[string]any, name string) (string, error) {
	var val string
	valI, ok := config[name]
	if !ok {
		return val, fmt.Errorf("checkNevents requires `%s` configuration parameter", name)
	}
	val, ok = valI.(string)
	if !ok {
		return val, fmt.Errorf("%s is not a string", name)
	}
	return val, nil
}

func GetOptionInt(config map[string]any, name string) (int, error) {
	var val int
	valI, ok := config[name]
	if !ok {
		return val, fmt.Errorf("checkNevents requires `%s` configuration parameter", name)
	}
	val, ok = valI.(int)
	if !ok {
		return val, fmt.Errorf("%s is not an int", name)
	}
	return val, nil
}

func GetOptionBool(config map[string]any, name string) (bool, error) {
	var val bool
	valI, ok := config[name]
	if !ok {
		return val, fmt.Errorf("checkNevents requires `%s` configuration parameter", name)
	}
	val, ok = valI.(bool)
	if !ok {
		return val, fmt.Errorf("%s is not an int", name)
	}
	return val, nil
}
