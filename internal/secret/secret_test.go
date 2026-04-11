package secret_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/secret"
)

func TestNewSecret(t *testing.T) {
	mb := &mockSecretBackend{}
	s := secret.NewSecret(mb, "default")
	require.NotNil(t, s)
	assert.Equal(t, secret.SecretAccessTokenType, s.Type)
	assert.NotNil(t, s.Revoke)
	assert.NotEmpty(t, s.Fields)
}
