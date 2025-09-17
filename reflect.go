package tools

import (
	"errors"
	"reflect"
)

var (
	ByteSliceType   = reflect.TypeOf([]byte{})
	StringType      = reflect.TypeOf("")
	StringSliceType = reflect.TypeOf([]string{})
	Int64Type       = reflect.TypeOf(int64(0))
	Int64SliceType  = reflect.TypeOf([]int64{})
)

func IsByteSlice(typ reflect.Type) bool {
	if typ.Kind() != reflect.Slice {
		return false
	}
	return typ.Elem().Kind() == reflect.Uint8
}

func SetByteSliceValue(val reflect.Value, bs []byte) error {
	if len(bs) > val.Cap() {
		if !val.CanSet() {
			return errors.New("tools: byte slice value cannot set")
		}
		val.Set(reflect.MakeSlice(val.Type(), len(bs), len(bs)))
	} else {
		val.SetLen(len(bs))
	}
	reflect.Copy(val, reflect.ValueOf(bs))
	return nil
}

func IsDefaultZero[T any](t T) bool {
	val := reflect.ValueOf(t)
	if !val.IsValid() {
		return true
	}
	return val.IsZero()
}

func IndirectType(typ reflect.Type) reflect.Type {
	for {
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		} else {
			return typ
		}
	}
}

func IndirectValue(val reflect.Value) reflect.Value {
	for {
		if val.Kind() == reflect.Pointer {
			val = val.Elem()
		} else {
			return val
		}
	}
}

func VariadicParam[T any](params []T, defaultValue ...T) T {
	var defaultVal T
	if len(defaultValue) > 0 {
		defaultVal = defaultValue[0]
	}
	if len(params) == 0 {
		return defaultVal
	}
	return params[0]
}
