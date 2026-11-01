package logging_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/logging"
)

func newLogger() (logging.Logger, *bytes.Buffer) {
	var out bytes.Buffer
	return logging.New(hclog.New(&hclog.LoggerOptions{Output: &out, Level: hclog.Trace})), &out
}

func TestTokenIsOnlyLoggedAsHash(t *testing.T) {
	log, out := newLogger()
	log.Debug("msg", logging.Fields{TokenID: 7, Token: "secret-token"})

	assert.NotContains(t, out.String(), "secret-token")
	assert.Contains(t, out.String(), "token_id=7")
	assert.Contains(t, out.String(), "token_sha1_hash="+logging.Hash("secret-token"))
}

func TestUnsetFieldsAreLeftOut(t *testing.T) {
	log, out := newLogger()
	log.Debug("msg", logging.Fields{ConfigName: "default"})

	assert.Contains(t, out.String(), "config_name=default")
	for _, key := range []string{"token_id", "token_sha1_hash", "error", "expires_at", "scopes"} {
		assert.NotContains(t, out.String(), key)
	}
}

func TestLevelsAndExtra(t *testing.T) {
	log, out := newLogger()
	log.Warn("warned", logging.Fields{Err: errors.New("boom")})
	log.Error("failed", logging.Fields{Extra: map[string]any{"provider": "token"}})
	log.Debug("no fields")
	log.Trace("traced", logging.Fields{ConfigName: "default"})
	log.Info("informed", logging.Fields{ConfigName: "default"})

	assert.Contains(t, out.String(), "[WARN]  warned: error=boom")
	assert.Contains(t, out.String(), "[ERROR] failed: provider=token")
	assert.Contains(t, out.String(), "[DEBUG] no fields")
	assert.Contains(t, out.String(), "[TRACE] traced: config_name=default")
	assert.Contains(t, out.String(), "[INFO]  informed: config_name=default")
}

func TestWithAddsFieldsToEveryLine(t *testing.T) {
	log, out := newLogger()
	log.With(logging.Fields{RoleName: "ci"}).Debug("msg", logging.Fields{Path: "group"})

	assert.Contains(t, out.String(), "role_name=ci")
	assert.Contains(t, out.String(), "path=group")
}
