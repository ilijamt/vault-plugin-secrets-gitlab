package utils

import (
	"fmt"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

// ConvertToInt attempts to convert various numeric types to an int.
//
// This function handles conversions from several numeric types (including int, int8,
// int16, int32, int64, float32, and float64) to a standard int. It uses type
// assertion to check the underlying type of the input. If the input is not a supported
// numeric type, it returns an error.
func ConvertToInt(num any) (int, error) {
	switch val := num.(type) {
	case int:
		return val, nil
	case int8:
		return int(val), nil
	case int16:
		return int(val), nil
	case int32:
		return int(val), nil
	case int64:
		return int(val), nil
	case float32:
		return int(val), nil
	case float64:
		return int(val), nil
	}
	return 0, fmt.Errorf("%v: %w", num, errs.ErrInvalidValue)
}

// ConvertToInt64 attempts to convert various numeric types to an int64.
//
// This function handles conversions from several numeric types (including int, int8,
// int16, int32, int64, float32, and float64) to a standard int. It uses type
// assertion to check the underlying type of the input. If the input is not a supported
// numeric type, it returns an error.
func ConvertToInt64(num any) (int64, error) {
	switch val := num.(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	}
	return int64(0), fmt.Errorf("%v: %w", num, errs.ErrInvalidValue)
}
