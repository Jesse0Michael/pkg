package boot

import "reflect"

// structHas checks if struct T contains an embedded field of type R and returns it
func structHas[R any, T any](t T) (R, bool) {
	var zero R
	v := reflect.ValueOf(t)
	typ := reflect.TypeOf(t)

	if v.Kind() != reflect.Struct {
		return zero, false
	}

	targetType := reflect.TypeOf(zero)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := typ.Field(i)

		// Check if field type matches R (either embedded or regular field)
		if fieldType.Type == targetType {
			return field.Interface().(R), true
		}
	}

	return zero, false
}
