package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/model/token"
)

func TestTokenPersonal(t *testing.T) {
	data := token.TokenPersonal{UserID: 1}
	assert.Contains(t, data.Data(), "user_id")
	assert.Contains(t, data.Event(nil), "user_id")
	assert.Contains(t, data.Internal(), "user_id")
	assert.EqualValues(t, 1, data.Data()["user_id"])
	assert.EqualValues(t, "1", data.Event(nil)["user_id"])
	assert.EqualValues(t, 1, data.Internal()["user_id"])
}
