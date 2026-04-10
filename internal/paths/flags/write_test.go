package flags_test

import (
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/flags"
	pathflags "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/flags"
)

func TestPathFlagsUpdate(t *testing.T) {
	f := flags.Flags{AllowRuntimeFlagsChange: true}
	mb := newMockFlagsBackend(t)

	// Flags() is called once during Paths() to check AllowRuntimeFlagsChange,
	// and once after the update to build the response.
	mb.MockFlagsProvider.EXPECT().Flags().Return(f).Once()
	mb.MockFlagsProvider.EXPECT().UpdateFlags(mock.Anything).Run(func(fn func(*flags.Flags)) {
		fn(&f)
	}).Return().Once()
	mb.MockEventSender.EXPECT().SendEvent(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	mb.MockFlagsProvider.EXPECT().Flags().Return(flags.Flags{
		ShowConfigToken:         true,
		AllowRuntimeFlagsChange: true,
	}).Once()

	p := pathflags.New(mb)
	paths := p.Paths()

	updateOp := paths[0].Operations[logical.UpdateOperation]
	require.NotNil(t, updateOp)

	fd := &framework.FieldData{
		Raw:    map[string]interface{}{"show_config_token": true},
		Schema: paths[0].Fields,
	}

	resp, err := updateOp.Handler()(t.Context(), &logical.Request{}, fd)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, true, resp.Data["show_config_token"])
	assert.Equal(t, true, resp.Data["allow_runtime_flags_change"])
	assert.True(t, f.ShowConfigToken)
}
