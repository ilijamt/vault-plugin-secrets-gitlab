package gitlab

import (
	"fmt"
	"time"
)

func allowedValues(values ...string) (ret []any) {
	for _, value := range values {
		ret = append(ret, value)
	}
	return ret
}

func convertToInt(num any) (int, error) {
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
	return 0, fmt.Errorf("%v: %w", num, ErrInvalidValue)
}

func calculateGitlabTTL(duration time.Duration, start time.Time) (ttl time.Duration, exp time.Time, err error) {
	const D = 24 * time.Hour
	var val = start.Add(duration).Round(0)
	exp = val.AddDate(0, 0, 1).Truncate(D)
	ttl = exp.Sub(start.Round(0))
	return ttl, exp, nil
}
