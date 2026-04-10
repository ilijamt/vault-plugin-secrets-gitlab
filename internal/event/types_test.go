package event_test

import (
	"testing"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/event"
	"github.com/stretchr/testify/require"
)

func TestEventType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "simple", input: "config-write"},
		{name: "single word", input: "write"},
		{name: "multiple hyphens", input: "static-creds-create-fail"},
		{name: "two parts", input: "revoke-access-token"},
		{name: "empty string", input: "", wantErr: true},
		{name: "uppercase", input: "Config-Write", wantErr: true},
		{name: "contains digits", input: "config1", wantErr: true},
		{name: "leading hyphen", input: "-write", wantErr: true},
		{name: "trailing hyphen", input: "write-", wantErr: true},
		{name: "double hyphen", input: "config--write", wantErr: true},
		{name: "contains space", input: "config write", wantErr: true},
		{name: "contains underscore", input: "config_write", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			et, err := event.NewEventType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.input, et.String())
			}
		})
	}
}

func TestMustEventType(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		require.NotPanics(t, func() {
			et := event.MustEventType("config-write")
			require.Equal(t, "config-write", et.String())
		})
	})

	t.Run("invalid panics", func(t *testing.T) {
		require.Panics(t, func() {
			event.MustEventType("INVALID")
		})
	})
}
