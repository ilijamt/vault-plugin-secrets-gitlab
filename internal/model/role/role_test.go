package role_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/role"
)

func TestRule(t *testing.T) {
	r := role.Role{Name: "Name"}
	require.False(t, r.IsNil())
	require.EqualValues(t, "Name", r.GetName())
	require.NotEmpty(t, r.LogicalResponseData())
}
