package main

import (
	"errors"
	"reflect"
)

var (
	errNotReferenceType = errors.New("out should be reference type")
	errCantSet          = errors.New("can't set value")
	errCantAssertData   = errors.New("can't assert data")
	errCantConvert      = errors.New("can't convert to needed type")
)

func i2s(data any, out any) error {
	return i2sReflectRecursive(data, reflect.ValueOf(out))
}

func i2sReflectRecursive(data any, out reflect.Value) error {

	switch out.Kind() {
	case reflect.Pointer, reflect.Slice:
	default:
		return errNotReferenceType
	}

	outValue := out.Elem()
	outValueType := outValue.Type()
	var toSet reflect.Value
	if !outValue.CanSet() {
		return errCantSet
	}

	switch outValue.Kind() {
	case reflect.Struct:
		dataMap, ok := (data).(map[string]any)
		if !ok {
			return errCantAssertData
		}

		toSet = reflect.New(outValue.Type()).Elem()
		for i := range outValue.NumField() {
			fieldName := outValue.Type().Field(i).Name
			fieldValuePtr := reflect.New(outValue.Field(i).Type())
			err := i2sReflectRecursive(dataMap[fieldName], fieldValuePtr)
			if err != nil {
				return err
			}
			toSet.Field(i).Set(fieldValuePtr.Elem())
		}

	case reflect.Slice:
		dataSlice, ok := (data).([]any)
		if !ok {
			return errCantAssertData
		}

		toSet = reflect.MakeSlice(outValue.Type(), len(dataSlice), cap(dataSlice))
		for i, dataEl := range dataSlice {
			elToFill := reflect.New(toSet.Type().Elem())
			err := i2sReflectRecursive(dataEl, elToFill)
			if err != nil {
				return err
			}
			toSet.Index(i).Set(elToFill.Elem())
		}

	default:
		if !reflect.TypeOf(data).ConvertibleTo(outValueType) {
			return errCantConvert
		}
		toSet = reflect.ValueOf(data).Convert(outValueType)
	}

	outValue.Set(toSet)
	return nil
}
