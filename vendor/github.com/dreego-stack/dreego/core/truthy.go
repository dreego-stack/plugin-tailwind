package core

import "reflect"

// Truthy reports whether v should render a template condition. It is the
// runtime backing of {#if}: nil, false, zero numbers, empty strings, and empty
// slices, maps, arrays, or channels are false; every other value is true.
// Values without a meaningful "empty" state, such as structs, stay true.
func Truthy(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return rv.Len() > 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	case reflect.Ptr, reflect.Interface, reflect.Func, reflect.UnsafePointer:
		return !rv.IsNil()
	default:
		return true
	}
}
