package utility

import "reflect"

// https://stackoverflow.com/questions/73956346/how-to-check-if-generic-type-is-nil
func IsNil[T any](t T) bool {
	v := reflect.ValueOf(t)
	kind := v.Kind()
	// Must be one of these types to be nillable
	return (kind == reflect.Ptr ||
		kind == reflect.Interface ||
		kind == reflect.Slice ||
		kind == reflect.Map ||
		kind == reflect.Chan ||
		kind == reflect.Func) &&
		v.IsNil()
}

func IsDefault[T comparable](arg T) bool {
	var t T
	return arg == t
}

func IsNilOrDefault[T comparable](arg T) bool {
	return IsNil(arg) || IsDefault(arg)
}

func IgnoreErr[T any](ret T, _ error) T { return ret }
