package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
)

func TestTokenGroupDeploy(t *testing.T) {
	data := token.TokenGroupDeploy{Username: "username"}
	assert.Contains(t, data.Data(), "username")
	assert.Contains(t, data.Event(nil), "username")
	assert.Contains(t, data.Internal(), "username")
	assert.EqualValues(t, "username", data.Data()["username"])
	assert.EqualValues(t, "username", data.Event(nil)["username"])
	assert.EqualValues(t, "username", data.Internal()["username"])
}
