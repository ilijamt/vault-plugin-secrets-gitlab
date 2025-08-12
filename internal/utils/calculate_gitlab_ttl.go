package utils

import (
	"time"
)

// CalculateGitlabTTL calculates the Time-To-Live (TTL) and expiration time for
// a GitLab-related operation based on a specified duration and start time.
//
// The function ensures that the calculated expiration does not exceed one year
// from the start time. It computes the expiration to be at the midnight
// following the calculated expiration date.
func CalculateGitlabTTL(duration time.Duration, start time.Time) (ttl time.Duration, exp time.Time, err error) {
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
