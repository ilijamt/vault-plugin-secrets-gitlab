package gitlab_test

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
)

func TestFlags_FlagSet(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	flags := &gitlab.Flags{}
	flags.FlagSet(fs)

	assert.False(t, flags.ShowConfigToken)
	assert.False(t, flags.AllowRuntimeFlagsChange)
	assert.NoError(t, fs.Parse([]string{"-show-config-token", "-allow-runtime-flags-change"}))
	assert.True(t, flags.ShowConfigToken)
	assert.True(t, flags.AllowRuntimeFlagsChange)
}
