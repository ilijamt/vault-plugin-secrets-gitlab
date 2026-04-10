package backend_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/backend"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
)

func TestInit(t *testing.T) {
	b := newTestBackend(t, flags.Flags{ShowConfigToken: true},
		backend.WithVersion("1.0.0"),
		backend.WithHelp("  help  "),
		backend.WithLocalStorage("local"),
		backend.WithSealWrapStorage("sealed"),
	)
	assert.Equal(t, "1.0.0", b.RunningVersion)
	assert.Equal(t, "help", b.Help)
	assert.Contains(t, b.PathsSpecial.LocalStorage, "local")
	assert.Contains(t, b.PathsSpecial.LocalStorage, framework.WALPrefix)
	assert.Contains(t, b.PathsSpecial.SealWrapStorage, "sealed")
	assert.True(t, b.Flags().ShowConfigToken)
}
