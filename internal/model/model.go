package model

// Model defines an interface for objects that contain a model and
// methods for collecting data relevant to model-based operations.
type Model interface {
	// GetName returns the model's name as a string
	GetName() string
	// LogicalResponseData returns a map containing relevant data that can be used in template operations or logical evaluations
	LogicalResponseData() map[string]any
	// IsNil returns a boolean indicating whether the instance is considered nil or invalid
	IsNil() bool
}
