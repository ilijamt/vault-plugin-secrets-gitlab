package utils_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

func TestCountErrByName(t *testing.T) {
	// Test with nil multierror
	var nilErr *multierror.Error
	result := utils.CountErrByName(nilErr)
	assert.Empty(t, result)

	// Test with multierror having nil Errors slice
	multiErr := &multierror.Error{Errors: nil}
	result = utils.CountErrByName(multiErr)
	assert.Empty(t, result)

	// Test with empty multierror
	emptyErr := &multierror.Error{Errors: []error{}}
	result = utils.CountErrByName(emptyErr)
	assert.Empty(t, result)

	// Test with nil error in slice
	multiErr = &multierror.Error{Errors: []error{nil}}
	result = utils.CountErrByName(multiErr)
	assert.Empty(t, result)

	// Test with single error
	err1 := fmt.Errorf("error1")
	multiErr = &multierror.Error{Errors: []error{err1}}
	result = utils.CountErrByName(multiErr)
	expected := map[string]int{"error1": 1}
	assert.Equal(t, expected, result)

	// Test with wrapped error
	wrappedErr := fmt.Errorf("wrapped: %w", fmt.Errorf("original"))
	multiErr = &multierror.Error{Errors: []error{wrappedErr}}
	result = utils.CountErrByName(multiErr)
	expected = map[string]int{"original": 1}
	assert.Equal(t, expected, result)

	// Test with non-wrapped error (Unwrap returns nil)
	nonWrappedErr := fmt.Errorf("non-wrapped")
	multiErr = &multierror.Error{Errors: []error{nonWrappedErr}}
	result = utils.CountErrByName(multiErr)
	expected = map[string]int{"non-wrapped": 1}
	assert.Equal(t, expected, result)

	// Test with multiple different errors
	err2 := fmt.Errorf("error2")
	multiErr = &multierror.Error{Errors: []error{err1, err2}}
	result = utils.CountErrByName(multiErr)
	expected = map[string]int{"error1": 1, "error2": 1}
	assert.Equal(t, expected, result)

	// Test with duplicate errors
	multiErr = &multierror.Error{Errors: []error{err1, err1, err2}}
	result = utils.CountErrByName(multiErr)
	expected = map[string]int{"error1": 2, "error2": 1}
	assert.Equal(t, expected, result)

	// Test with mix of nil and valid errors
	multiErr = &multierror.Error{Errors: []error{nil, err1, nil, err2, err1}}
	result = utils.CountErrByName(multiErr)
	expected = map[string]int{"error1": 2, "error2": 1}
	assert.Equal(t, expected, result)

	// Test with wrapped and non-wrapped errors mixed
	multiErr = &multierror.Error{Errors: []error{wrappedErr, nonWrappedErr}}
	result = utils.CountErrByName(multiErr)
	expected = map[string]int{"original": 1, "non-wrapped": 1}
	assert.Equal(t, expected, result)
}
