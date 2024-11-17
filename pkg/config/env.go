package config

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

const TagKey = "env"

type Env struct {
	source func(string) string
}

func NewEnv(source func(string) string) *Env {
	return &Env{
		source: source,
	}
}

func (e *Env) Set(conf any) error {
	refConf := reflect.ValueOf(conf)
	if refConf.Kind() != reflect.Ptr && refConf.Kind() != reflect.Interface {
		return errors.New("conf must be a pointer")
	}

	// Extract the value from the pointer or empty interface{}
	refConfStruct, err := e.extractConcreteStruct(refConf)
	if err != nil {
		return errors.New("could not extract concrete struct")
	}

	if refConfStruct.Kind() != reflect.Struct {
		return errors.New("conf must be a struct")
	}

	err = e.fillStruct(refConfStruct)
	if err != nil {
		return err
	}

	return nil
}

func (e *Env) extractConcreteStruct(refConf reflect.Value) (reflect.Value, error) {
	if refConf.Kind() != reflect.Ptr && refConf.Kind() != reflect.Interface {
		if refConf.Kind() != reflect.Struct {
			return reflect.Value{}, errors.New("conf must be a struct")
		}

		return refConf, nil
	}

	return e.extractConcreteStruct(refConf.Elem())
}

func (e *Env) fillStruct(value reflect.Value) error {
	for i := 0; i < value.NumField(); i++ {
		valueField := value.Field(i)
		valueFieldType := value.Type().Field(i)

		if !valueFieldType.IsExported() {
			return fmt.Errorf("unexported field type: %s", valueFieldType.Type)
		}

		switch valueField.Kind() {
		case reflect.Struct:
			err := e.fillStruct(valueField)
			if err != nil {
				return err
			}
		case reflect.String:
			valueField.SetString(e.source(valueFieldType.Tag.Get(TagKey)))
		case reflect.Bool:
			v, err := strconv.ParseBool(e.source(valueFieldType.Tag.Get(TagKey)))
			if err != nil {
				return err
			}

			valueField.SetBool(v)
		case reflect.Int:
			v, err := strconv.ParseInt(e.source(valueFieldType.Tag.Get(TagKey)), 10, 64)
			if err != nil {
				return err
			}

			valueField.SetInt(v)
		case reflect.Float32, reflect.Float64:
			v, err := strconv.ParseFloat(e.source(valueFieldType.Tag.Get(TagKey)), 64)
			if err != nil {
				return err
			}

			valueField.SetFloat(v)
		default:
			return fmt.Errorf("unsupported type %s for field %s", valueField.Kind(), valueFieldType.Type)
		}
	}

	return nil
}
