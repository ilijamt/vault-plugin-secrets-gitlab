package utils

import (
	"fmt"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

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
