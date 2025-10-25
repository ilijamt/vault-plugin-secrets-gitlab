package config

import "time"

const (
	DefaultAutoRotateBeforeMinTTL = 24 * time.Hour
	DefaultAutoRotateBeforeMaxTTL = 730 * time.Hour
)
