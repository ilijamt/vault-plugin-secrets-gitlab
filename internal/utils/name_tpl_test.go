package utils_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

type tokenName struct {
	name string
	data map[string]any
}

func (t *tokenName) IsNil() bool {
	return t == nil
}

func (t *tokenName) GetName() string {
	return t.name
}

func (t *tokenName) LogicalResponseData() map[string]any {
	if t.data == nil {
		return make(map[string]any)
	}
	return t.data
}

var _ utils.TokenNameData = (*tokenName)(nil)

func TestTokenNameGenerator(t *testing.T) {
	var tests = []struct {
		in     *tokenName
		outVal string
		outErr bool
	}{
		{nil, "", true},

		// invalid template
		{
			&tokenName{
				name: "{{ .role_name",
			},
			"",
			true,
		},

		// combination template
		{
			&tokenName{
				name: "{{ .role_name }}-{{ .token_type }}-access-token-{{ yesNoBool .gitlab_revokes_token }}",
				data: map[string]any{
					"role_name":            "test",
					"token_type":           "personal",
					"gitlab_revokes_token": true,
				},
			},
			"test-personal-access-token-yes",
			false,
		},

		// with stringsJoin
		{
			&tokenName{
				name: "{{ .role_name }}-{{ .token_type }}-{{ stringsJoin .scopes \"-\" }}-{{ yesNoBool .gitlab_revokes_token }}",
				data: map[string]any{
					"role_name":            "test",
					"token_type":           "personal",
					"scopes":               []string{"api", "sudo"},
					"gitlab_revokes_token": false,
				},
			},
			"test-personal-api-sudo-no",
			false,
		},

		// with timeNowFormat
		{
			&tokenName{
				name: "{{ .role_name }}-{{ .token_type }}-{{ timeNowFormat \"2006-01\" }}",
				data: map[string]any{
					"role_name":  "test",
					"token_type": "personal",
				},
			},
			fmt.Sprintf("test-personal-%d-%02d", time.Now().UTC().Year(), time.Now().UTC().Month()),
			false,
		},
	}

	for _, tst := range tests {
		t.Logf("TokenName(%v)", tst.in)
		val, err := utils.TokenName(tst.in)
		assert.Equal(t, tst.outVal, val)
		if tst.outErr {
			assert.Error(t, err, tst.outErr)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestValidateTokenNameName(t *testing.T) {
	require.Error(t, utils.ValidateTokenNameName(&tokenName{name: "{{ .name"}))
	require.Error(t, utils.ValidateTokenNameName(nil))
}

func TestTokenNameGenerator_RandString(t *testing.T) {
	val, err := utils.TokenName(
		&tokenName{
			name: "{{ randHexString 8 }}",
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, val)
	require.Len(t, val, 16)
}

func TestTokenNameGenerator_UnixTimeStamp(t *testing.T) {
	now := time.Now().UTC().Unix()
	val, err := utils.TokenName(
		&tokenName{
			name: "{{ .unix_timestamp_utc }}",
		},
	)
	require.NoError(t, err)
	require.NotEmpty(t, val)
	i, err := strconv.ParseInt(val, 10, 64)
	require.NoError(t, err)
	require.GreaterOrEqual(t, i, now)
}
