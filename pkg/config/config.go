package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

func LoadConfig(config interface{}, path string) error {
	err := parseDefault(config)
	if err != nil {
		return err
	}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		_ = yaml.Unmarshal(content, config)
	} else if strings.HasSuffix(path, ".json") {
		_ = json.Unmarshal(content, config)
	} else {
		return errors.New(fmt.Sprintf("unsupported config type %v", path))
	}
	err = verifyConfig(config)
	if err != nil {
		return err
	}
	return nil
}

func parseDefault(config interface{}) error {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	if configValue.Kind() != reflect.Struct {
		return errors.New("invalid config, should be struct")
	}

	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {
		var (
			fieldStruct = configType.Field(i)
			field       = configValue.Field(i)
		)

		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		if isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()); isBlank {
			if value := fieldStruct.Tag.Get("default"); value != "" {
				if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
					return err
				}
			}
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.Struct:
			if err := parseDefault(field.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Slice:
			for i := 0; i < field.Len(); i++ {
				if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
					if err := parseDefault(field.Index(i).Addr().Interface()); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func verifyConfig(config interface{}) error {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {
		var (
			fieldStruct = configType.Field(i)
			field       = configValue.Field(i)
		)

		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		if isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()); isBlank {
			if value := fieldStruct.Tag.Get("required"); value != "" {
				return errors.New(fmt.Sprintf("field %v is required", fieldStruct.Name))
			}
		}

		if field.Type().Comparable() {
			if value := fieldStruct.Tag.Get("max"); value != "" {
				gt, err := compare(field, value, true)
				if err != nil {
					return err
				}
				if !gt {
					return errors.New(fmt.Sprintf("field [%s] = %v > max %s", fieldStruct.Name, field, value))
				}
			} else if value := fieldStruct.Tag.Get("min"); value != "" {
				lt, err := compare(field, value, false)
				if err != nil {
					return err
				}
				if !lt {
					return errors.New(fmt.Sprintf("field [%s] = %v < min %s", fieldStruct.Name, field, value))
				}
			}
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.Struct:
			if err := verifyConfig(field.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Slice:
			for i := 0; i < field.Len(); i++ {
				if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
					if err := verifyConfig(field.Index(i).Addr().Interface()); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func compare(t reflect.Value, value string, gt bool) (bool, error) {
	if t.CanInt() {
		vInt, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return false, errors.New(fmt.Sprintf("value %v can't be conv to int", value))
		}
		return ((t.Int() > vInt) != gt) || vInt == t.Int(), nil
	} else if t.CanUint() {
		vUint, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return false, errors.New(fmt.Sprintf("value %v can't be conv to uint", value))
		}
		return (t.Uint() > vUint) != gt || vUint == t.Uint(), nil
	} else if t.CanFloat() {
		vFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return false, errors.New(fmt.Sprintf("value %v can't be conv to float", value))
		}
		return (t.Float() > vFloat) != gt || vFloat == t.Float(), nil
	}
	return false, errors.New("not number type")
}
