package gitlab

import "fmt"

func allowedValues(values ...string) (ret []interface{}) {
	for _, value := range values {
		ret = append(ret, value)
	}
	return ret
}

func convertToInt(num any) (int, error) {
	switch val := num.(type) {
	case int:
		return val, nil
	case float64:
		return int(val), nil
	}
	return 0, fmt.Errorf("%v: %w", num, ErrInvalidValue)
}
