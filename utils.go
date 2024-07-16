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
	start = start.UTC()
	const D = 24 * time.Hour
	const maxDuration = 365 * 24 * time.Hour
	if duration > maxDuration {
		duration = maxDuration
	}
	var val = start.Add(duration).Round(0)
	exp = val.AddDate(0, 0, 1).Truncate(D)
	ttl = exp.Sub(start.Round(0))
	if ttl > maxDuration {
		m := start.Add(maxDuration)
		exp = time.Date(m.Year(), m.Month(), m.Day(), 0, 0, 0, 0, m.Location())
		ttl = exp.Sub(start.Round(0))
	}
	return ttl, exp, nil
}
