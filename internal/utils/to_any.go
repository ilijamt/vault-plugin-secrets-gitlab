package utils

// ToAny converts a slice of values of type int or string to a slice of empty interfaces (any).
//
// This function is a generic utility that allows for the conversion of a variadic list
// of integers or strings into a slice of empty interfaces (`[]any`). This is particularly
// useful when you need to pass mixed-type values around, or when API requirements dictate
// an `any` type.
func ToAny[T int | string](values ...T) (ret []any) {
	ret = make([]any, 0, len(values))
	for _, value := range values {
		ret = append(ret, value)
	}
	return ret
}
