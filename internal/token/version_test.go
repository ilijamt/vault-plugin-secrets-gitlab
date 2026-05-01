package token_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
)

// Version semantics are not directly exported; they are exercised through
// IsScopeAllowed using ScopeSelfRotate (since 17.9), the most version-sensitive
// boundary on a Personal token.
func TestVersionGating_BoundaryConditions(t *testing.T) {
	const since = "17.9"
	_ = since // documents the boundary under test

	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{"empty version is lenient", "", true},
		{"only major below boundary", "17", false},
		{"exact boundary minor", "17.9", true},
		{"minor above boundary", "17.10", true},
		{"major above boundary", "18.0", true},
		{"patch + pre-release suffix above boundary", "17.9.1-pre", true},
		{"patch + ee suffix below boundary", "17.8.5-ee", false},
		{"patch + ee suffix above boundary", "18.0.0-ee", true},
		{"v-prefix major above", "v18", true},
		{"unparseable version is lenient", "not-a-version", true},
		{"in-memory test stub literal 'version' is lenient", "version", true},
		{"major below boundary", "16.4", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := token.IsScopeAllowed(token.TypePersonal, token.ScopeSelfRotate, tc.version)
			assert.Equal(t, tc.want, got)
		})
	}
}
