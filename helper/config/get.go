package config

import (
	"errors"

	"github.com/ansel1/merry/v2"
)

var (
	ErrUnknownConfigKey = errors.New("unknown configuration key")
	ErrNotAString       = errors.New("value is not a string")
	ErrNotAnInt         = errors.New("value is not an int")
	ErrNotABool         = errors.New("value is not a bool")
)

func GetOptionString(config map[string]any, name string) (string, error) {
	var val string
	valI, ok := config[name]
	if !ok {
		return val, merry.Wrap(ErrUnknownConfigKey, merry.WithMessagef("'%s' argument must be specified", name))
	}
	val, ok = valI.(string)
	if !ok {
		return val, merry.Wrap(ErrNotAString, merry.WithMessagef("%s is not a string", name))
	}
	return val, nil
}

func GetOptionStringWithDefault(config map[string]any, name string, def string) string {
	val, err := GetOptionString(config, name)
	if err != nil {
		return def
	}
	return val
}

func GetOptionInt(config map[string]any, name string) (int, error) {
	var val int
	valI, ok := config[name]
	if !ok {
		return val, merry.Wrap(ErrUnknownConfigKey, merry.WithMessagef("'%s' argument must be specified", name))
	}
	val, ok = valI.(int)
	if !ok {
		return val, merry.Wrap(ErrNotAnInt, merry.WithMessagef("%s is not an int", name))
	}
	return val, nil
}

func GetOptionBool(config map[string]any, name string) (bool, error) {
	var val bool
	valI, ok := config[name]
	if !ok {
		return val, merry.Wrap(ErrUnknownConfigKey, merry.WithMessagef("'%s' argument must be specified", name))
	}
	val, ok = valI.(bool)
	if !ok {
		return val, merry.Wrap(ErrNotABool, merry.WithMessagef("%s is not a bool", name))
	}
	return val, nil
}

func GetOptionBoolWithDefault(config map[string]any, name string, def bool) (bool, error) {
	var val bool
	valI, ok := config[name]
	if !ok {
		return def, nil
	}
	val, ok = valI.(bool)
	if !ok {
		return def, merry.Wrap(ErrNotABool, merry.WithMessagef("%s is not a bool", name))
	}
	return val, nil
}
