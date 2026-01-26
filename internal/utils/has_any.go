package utils

// HasAny returns true if pred(s, v) is true for any v in vals.
func HasAny[T any](s T, vals []T, pred func(T, T) bool) bool {
	for _, v := range vals {
		if pred(s, v) {
			return true
		}
	}
	return false
}
