package utils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestToAny(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []any
	}{
		{
			name:     "empty int slice",
			input:    []int{},
			expected: []any{},
		},
		{
			name:     "single int",
			input:    []int{42},
			expected: []any{42},
		},
		{
			name:     "multiple ints",
			input:    []int{1, 2, 3},
			expected: []any{1, 2, 3},
		},
		{
			name:     "empty string slice",
			input:    []string{},
			expected: []any{},
		},
		{
			name:     "single string",
			input:    []string{"hello"},
			expected: []any{"hello"},
		},
		{
			name:     "multiple strings",
			input:    []string{"a", "b", "c"},
			expected: []any{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []any
			switch v := tt.input.(type) {
			case []int:
				result = utils.ToAny(v...)
			case []string:
				result = utils.ToAny(v...)
			}

			assert.EqualValues(t, tt.expected, result)
		})
	}
}
