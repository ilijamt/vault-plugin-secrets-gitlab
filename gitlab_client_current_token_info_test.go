package gitlab

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xanzy/go-gitlab"
)

// mockGitlabClient is a stub implementation that simulates the GitLab client for testing
type mockGitlabClient struct {
	gitlabClient
	personalAccessTokensFunc func(ctx context.Context) ([]*gitlab.PersonalAccessToken, error)
}

func (m *mockGitlabClient) PersonalAccessTokens(ctx context.Context) ([]*gitlab.PersonalAccessToken, error) {
	if m.personalAccessTokensFunc != nil {
		return m.personalAccessTokensFunc(ctx)
	}
	return nil, nil
}

// TestCurrentTokenInfo_WithExpiryDate tests that tokens with expiry dates are not modified
func TestCurrentTokenInfo_WithExpiryDate(t *testing.T) {
	// Create a timestamp for testing
	now := time.Now()
	expiryDate := now.AddDate(0, 2, 0) // 2 months in the future
	
	// Create a mock client
	mockClient := &mockGitlabClient{
		personalAccessTokensFunc: func(ctx context.Context) ([]*gitlab.PersonalAccessToken, error) {
			return []*gitlab.PersonalAccessToken{
				{
					ID:        123,
					Name:      "Test Token",
					CreatedAt: &now,
					ExpiresAt: &expiryDate,
					UserID:    456,
				},
			}, nil
		},
	}
	
	// Call the method being tested
	tokenInfo, err := mockClient.CurrentTokenInfo(context.Background())
	
	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, tokenInfo)
	require.NotNil(t, tokenInfo.ExpiresAt)
	assert.Equal(t, expiryDate, *tokenInfo.ExpiresAt, "Expiry date should not be modified")
}

// TestCurrentTokenInfo_WithoutExpiryDate tests that tokens without expiry dates get an artificial one
func TestCurrentTokenInfo_WithoutExpiryDate(t *testing.T) {
	// Create a timestamp for testing
	now := time.Now()
	expectedExpiry := now.AddDate(1, 0, -2) // 1 year minus 2 days from creation
	
	// Create a mock client
	mockClient := &mockGitlabClient{
		personalAccessTokensFunc: func(ctx context.Context) ([]*gitlab.PersonalAccessToken, error) {
			return []*gitlab.PersonalAccessToken{
				{
					ID:        123,
					Name:      "Test Token",
					CreatedAt: &now,
					ExpiresAt: nil, // No expiry date set
					UserID:    456,
				},
			}, nil
		},
	}
	
	// Call the method being tested
	tokenInfo, err := mockClient.CurrentTokenInfo(context.Background())
	
	// Verify the result
	require.NoError(t, err)
	require.NotNil(t, tokenInfo)
	require.NotNil(t, tokenInfo.ExpiresAt)
	
	// Calculate the difference between expected and actual times
	// We need to allow for a small difference due to execution time
	timeDiff := expectedExpiry.Sub(*tokenInfo.ExpiresAt)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	
	// Verify the expiry date was set correctly (within a small tolerance)
	assert.True(t, timeDiff < time.Second, "Expiry date should be set to 1 year minus 2 days from creation")
}