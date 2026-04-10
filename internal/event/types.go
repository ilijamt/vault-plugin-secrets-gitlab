package event

import (
	"fmt"
	"regexp"
)

// validName matches lowercase letters and hyphens, e.g. "config-write", "token-create".
var validName = regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)

// EventType represents an event type name like "config-write" or "static-creds-create-fail".
type EventType string

// NewEventType creates a validated EventType.
// The name must contain only lowercase letters (a-z) and hyphens.
func NewEventType(name string) (EventType, error) {
	if !validName.MatchString(name) {
		return "", fmt.Errorf("invalid event type: %q", name)
	}
	return EventType(name), nil
}

// MustEventType is like NewEventType but panics on invalid input.
func MustEventType(name string) EventType {
	et, err := NewEventType(name)
	if err != nil {
		panic(err)
	}
	return et
}

func (e EventType) String() string {
	return string(e)
}
