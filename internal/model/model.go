package model

// ExtendedModel defines an interface for objects that contain a model and
// methods for collecting data relevant to model-based operations.
type ExtendedModel interface {
	Named
	IsNil
	Event
	LogicalResponseData
}

type Named interface {
	// GetName returns the model's name as a string
	GetName() string
}

type IsNil interface {
	// IsNil returns a boolean indicating whether the instance is considered nil or invalid
	IsNil() bool
}

type Event interface{}

type LogicalResponseData interface {
	// LogicalResponseData returns a map containing relevant data that can be used in template operations or logical evaluations
	LogicalResponseData() map[string]any
}
