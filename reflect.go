package tools

import (
	"errors"
	"reflect"
)

var (
	// ByteSliceType is the reflect.Type for []byte.
	ByteSliceType = reflect.TypeOf([]byte{})
	// StringType is the reflect.Type for string.
	StringType = reflect.TypeOf("")
	// StringSliceType is the reflect.Type for []string.
	StringSliceType = reflect.TypeOf([]string{})
	// Int64Type is the reflect.Type for int64.
	Int64Type = reflect.TypeOf(int64(0))
	// Int64SliceType is the reflect.Type for []int64.
	Int64SliceType = reflect.TypeOf([]int64{})
)

// IsByteSlice reports whether the given reflect.Type represents a byte slice ([]byte).
func IsByteSlice(typ reflect.Type) bool {
	if typ.Kind() != reflect.Slice {
		return false
	}
	return typ.Elem().Kind() == reflect.Uint8
}

// SetByteSliceValue sets the reflect.Value val (which must represent a []byte) to
// the contents of bs. If val's capacity is insufficient, it allocates a new slice
// (val must be settable in that case). Returns an error if val cannot be set.
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

// IsDefaultZero reports whether t is the zero value of its type.
// An invalid reflect.Value is also considered zero.
func IsDefaultZero[T any](t T) bool {
	val := reflect.ValueOf(t)
	if !val.IsValid() {
		return true
	}
	return val.IsZero()
}

// IndirectType dereferences typ until it reaches a non-pointer type.
// If typ is not a pointer, it is returned unchanged.
func IndirectType(typ reflect.Type) reflect.Type {
	for {
		if typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		} else {
			return typ
		}
	}
}

// IndirectValue dereferences val until it reaches a non-pointer value.
// If val is not a pointer, it is returned unchanged.
func IndirectValue(val reflect.Value) reflect.Value {
	for {
		if val.Kind() == reflect.Pointer {
			val = val.Elem()
		} else {
			return val
		}
	}
}

// IsNil reports whether v is nil. It handles both untyped nil and typed nil
// values of kinds Chan, Func, Interface, Map, Ptr, and Slice. An invalid
// reflect.Value is also considered nil.
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return true
	}
	switch val.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return val.IsNil()
	default:
		return false
	}
}
