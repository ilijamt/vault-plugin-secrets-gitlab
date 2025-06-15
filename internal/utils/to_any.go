package utils

func ToAny[T int | string](values ...T) (ret []any) {
	for _, value := range values {
		ret = append(ret, value)
	}
	return ret
}
