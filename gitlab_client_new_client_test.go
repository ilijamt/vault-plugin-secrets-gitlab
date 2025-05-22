package gitlab

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewGitlabClient_EmptyBaseURL tests that newGitlabClient properly validates an empty BaseURL
func TestNewGitlabClient_EmptyBaseURL(t *testing.T) {
	config := &EntryConfig{
		BaseURL: "",           // Empty base URL
		Token:   "valid-token", // Valid token
	}
	
	// Call the function being tested
	client, err := newGitlabClient(config, http.DefaultClient)
	
	// Verify the result
	assert.Nil(t, client, "Client should be nil due to invalid base URL")
	assert.Error(t, err, "Should return an error for empty base URL")
	assert.ErrorIs(t, err, ErrInvalidValue, "Error should include ErrInvalidValue")
}

// TestNewGitlabClient_WhitespaceBaseURL tests that newGitlabClient properly validates a whitespace-only BaseURL
func TestNewGitlabClient_WhitespaceBaseURL(t *testing.T) {
	config := &EntryConfig{
		BaseURL: "   ",         // Whitespace-only base URL
		Token:   "valid-token", // Valid token
	}
	
	// Call the function being tested
	client, err := newGitlabClient(config, http.DefaultClient)
	
	// Verify the result
	assert.Nil(t, client, "Client should be nil due to invalid base URL")
	assert.Error(t, err, "Should return an error for whitespace-only base URL")
	assert.ErrorIs(t, err, ErrInvalidValue, "Error should include ErrInvalidValue")
}

// TestNewGitlabClient_EmptyToken tests that newGitlabClient properly validates an empty Token
func TestNewGitlabClient_EmptyToken(t *testing.T) {
	config := &EntryConfig{
		BaseURL: "https://gitlab.example.com", // Valid base URL
		Token:   "",                          // Empty token
	}
	
	// Call the function being tested
	client, err := newGitlabClient(config, http.DefaultClient)
	
	// Verify the result
	assert.Nil(t, client, "Client should be nil due to invalid token")
	assert.Error(t, err, "Should return an error for empty token")
	assert.ErrorIs(t, err, ErrInvalidValue, "Error should include ErrInvalidValue")
}

// TestNewGitlabClient_WhitespaceToken tests that newGitlabClient properly validates a whitespace-only Token
func TestNewGitlabClient_WhitespaceToken(t *testing.T) {
	config := &EntryConfig{
		BaseURL: "https://gitlab.example.com", // Valid base URL
		Token:   "   ",                       // Whitespace-only token
	}
	
	// Call the function being tested
	client, err := newGitlabClient(config, http.DefaultClient)
	
	// Verify the result
	assert.Nil(t, client, "Client should be nil due to invalid token")
	assert.Error(t, err, "Should return an error for whitespace-only token")
	assert.ErrorIs(t, err, ErrInvalidValue, "Error should include ErrInvalidValue")
}