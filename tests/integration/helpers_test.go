//go:build paths || saas || serviceaccount || e2e

package integration_test

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/require"
	g "gitlab.com/gitlab-org/api/client-go/v2"
	"golang.org/x/mod/semver"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"

	gitlab "github.com/ilijamt/vault-plugin-secrets-gitlab"
	gitlabTypes "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/gitlab/types"
	tokenPaths "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/paths/token"
	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

var (
	gitlabComPersonalAccessToken = cmp.Or(os.Getenv("GITLAB_COM_TOKEN"), "glpat-invalid-value")
	gitlabComUrl                 = cmp.Or(os.Getenv("GITLAB_COM_URL"), "https://gitlab.com")
)

// serviceAccountMinVersion is the first GitLab version where service accounts
// (user, group and project) are available on CE/Free; before 18.11 they were
// Premium+, so the API 404s on CE.
const serviceAccountMinVersion = "18.11"

// requireServiceAccounts skips the test when GITLAB_VERSION predates CE service
// account support.
func requireServiceAccounts(tb testing.TB) {
	tb.Helper()
	if !gitlabVersionAtLeast(os.Getenv("GITLAB_VERSION"), serviceAccountMinVersion) {
		tb.Skipf("service accounts require GitLab CE >= %s", serviceAccountMinVersion)
	}
}

// gitlabVersionAtLeast reports whether version >= minVersion, leniently for unparseable versions.
func gitlabVersionAtLeast(version, minVersion string) bool {
	canon := func(v string) string {
		if !strings.HasPrefix(v, "v") {
			v = "v" + v
		}
		return v
	}
	a, b := canon(version), canon(minVersion)
	if !semver.IsValid(a) {
		return true
	}
	return semver.Compare(a, b) >= 0
}

// validScopesFor returns every scope the given token type accepts at the
// GitLab version reported via GITLAB_VERSION (set by the integration test
// harness via `make test`). When GITLAB_VERSION is unset the gate is lenient
// and every known scope is returned. Returns nil for token types that do not
// take a scopes field (e.g. pipeline trigger).
func validScopesFor(tokenType t.Type) []string {
	scopes, applicable := t.ValidScopesFor(tokenType, os.Getenv("GITLAB_VERSION"))
	if !applicable {
		return nil
	}
	out := make([]string, 0, len(scopes))
	for _, s := range scopes {
		out = append(out, s.String())
	}
	return out
}

func sanitizePath(path string) string {
	var builder strings.Builder

	for _, r := range path {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}

	return strings.ReplaceAll(builder.String(), "__", "_")
}

// cassettePath returns the cassette name (without .yaml) for the test; "paths",
// "e2e" and "serviceaccount" are pinned to a GitLab version, "saas" is not.
func cassettePath(tb testing.TB, target string) string {
	tb.Helper()
	switch target {
	case "paths", "e2e", "serviceaccount":
		version := os.Getenv("GITLAB_VERSION")
		if version == "" {
			tb.Fatal("GITLAB_VERSION env var must be set for paths/e2e/serviceaccount cassettes; run via 'make test' or export GITLAB_VERSION explicitly")
		}
		return fmt.Sprintf("testdata/%s/%s/%s", target, version, sanitizePath(tb.Name()))
	default:
		return fmt.Sprintf("testdata/%s/%s", target, sanitizePath(tb.Name()))
	}
}

func getCtxGitlabClient(t *testing.T, target string) context.Context {
	t.Helper()
	httpClient, _ := getClient(t, target)
	return utils.HttpClientNewContext(t.Context(), httpClient)
}

func getCtxGitlabClientWithUrl(t *testing.T, target string) (context.Context, string) {
	t.Helper()
	httpClient, url := getClient(t, target)
	return utils.HttpClientNewContext(t.Context(), httpClient), url
}

// cassetteTime reads the test's static clock from its own cassette: the config
// token's created_at (GET /personal_access_tokens/self), else the expires_at in
// the first recorded request (direct client tests), else the zero time. A
// missing cassette means we are recording, so it falls back to the wall clock.
func cassetteTime(tb testing.TB, target string) time.Time {
	tb.Helper()
	c, err := cassette.Load(cassettePath(tb, target))
	if err != nil {
		return time.Now()
	}
	for _, i := range c.Interactions {
		if !strings.HasSuffix(i.Request.URL, "/personal_access_tokens/self") {
			continue
		}
		var body struct {
			CreatedAt time.Time `json:"created_at"`
		}
		if err := json.Unmarshal([]byte(i.Response.Body), &body); err == nil {
			return body.CreatedAt
		}
		return time.Time{}
	}
	for _, i := range c.Interactions {
		var body struct {
			ExpiresAt string `json:"expires_at"`
		}
		if json.Unmarshal([]byte(i.Request.Body), &body) == nil && body.ExpiresAt != "" {
			if t, err := time.Parse(time.DateOnly, body.ExpiresAt); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func ctxTestTime(ctx context.Context, tb testing.TB, target string) (context.Context, time.Time) {
	tb.Helper()
	t := cassetteTime(tb, target)
	return utils.WithStaticTime(ctx, t), t
}

func filterSlice[T any, Slice ~[]T](collection Slice, predicate func(item T, index int64) bool) Slice {
	result := make(Slice, 0, len(collection))

	for i := range collection {
		if predicate(collection[i], int64(i)) {
			result = append(result, collection[i])
		}
	}

	return result
}

type generatedToken struct {
	Token string `json:"token"`
}

type generatedTokens map[string]generatedToken

// placeholderToken is the config token used on replay when no recorded token set
// is present; the matcher ignores auth headers, so it only needs to be non-empty
// and (with the name appended) distinct per token so SHAs differ.
const placeholderToken = "REPLACED-TOKEN"

// loadTokens reads the recording-only token set (testdata/tokens.<version>.json);
// when it is absent, getGitlabToken falls back to placeholderToken.
var loadTokens = sync.OnceValues(func() (generatedTokens, error) {
	version := os.Getenv("GITLAB_VERSION")
	if version == "" {
		return generatedTokens{}, nil
	}
	payload, err := os.ReadFile(fmt.Sprintf("./testdata/tokens.%s.json", version))
	if errors.Is(err, os.ErrNotExist) {
		return generatedTokens{}, nil
	}
	if err != nil {
		return nil, err
	}
	var t generatedTokens
	err = json.Unmarshal(payload, &t)
	return t, err
})

func getGitlabToken(name string) generatedToken {
	tokens, _ := loadTokens()
	if token, ok := tokens[name]; ok && token.Token != "" {
		return token
	}
	return generatedToken{Token: placeholderToken + "-" + name}
}

// standardConfig is the self-rotating backend config shared by the flow tests.
func standardConfig(typ gitlabTypes.Type, url, token string) map[string]any {
	return map[string]any{
		"token":              token,
		"base_url":           url,
		"auto_rotate_token":  true,
		"auto_rotate_before": "24h",
		"type":               typ.String(),
	}
}

// issueToken reads a token from roleName and returns it with its lease secret.
// The clock is pinned to the cassette so expiry is deterministic on replay.
func issueToken(ctx context.Context, tb testing.TB, b *gitlab.Backend, l logical.Storage, target, roleName string) (string, *logical.Secret) {
	tb.Helper()
	ctxIssue, _ := ctxTestTime(ctx, tb, target)
	resp, err := b.HandleRequest(ctxIssue, &logical.Request{
		Operation: logical.ReadOperation, Storage: l,
		Path: fmt.Sprintf("%s/%s", tokenPaths.PathTokenRoleStorage, roleName),
	})
	require.NoError(tb, err)
	require.NotNil(tb, resp)
	require.NoError(tb, resp.Error())
	token, ok := resp.Data["token"].(string)
	require.True(tb, ok)
	require.NotEmpty(tb, token)
	require.NotNil(tb, resp.Secret)
	return token, resp.Secret
}

func revokeSecret(ctx context.Context, tb testing.TB, b *gitlab.Backend, l logical.Storage, secret *logical.Secret) {
	tb.Helper()
	resp, err := b.HandleRequest(ctx, &logical.Request{
		Operation: logical.RevokeOperation,
		Path:      "/",
		Storage:   l,
		Secret:    secret,
	})
	require.NoError(tb, err)
	require.Nil(tb, resp)
}

// requireTokenStatus asserts GET /personal_access_tokens/self with token yields
// wantStatus: StatusOK while the token is live, StatusUnauthorized once revoked.
func requireTokenStatus(tb testing.TB, httpClient *http.Client, url, token string, wantStatus int) {
	tb.Helper()
	c, err := g.NewClient(token, g.WithHTTPClient(httpClient), g.WithBaseURL(url))
	require.NoError(tb, err)
	require.NotNil(tb, c)
	pat, r, err := c.PersonalAccessTokens.GetSinglePersonalAccessToken()
	require.NotNil(tb, r)
	require.EqualValues(tb, wantStatus, r.StatusCode)
	if wantStatus == http.StatusOK {
		require.NoError(tb, err)
		require.NotNil(tb, pat)
	} else {
		require.Error(tb, err)
		require.Nil(tb, pat)
	}
}
