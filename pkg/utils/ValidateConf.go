package utils

import (
	"errors"
	"reflect"
)

func ValidateConf(conf interface{}) error {
	v := reflect.ValueOf(conf)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == "" {
			return errors.New("Config: " + v.Type().Field(i).Name + " empty!")
		}
	}
	return nil
}
