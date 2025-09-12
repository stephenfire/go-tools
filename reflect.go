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
			return errors.New("byte slice value cannot set")
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
