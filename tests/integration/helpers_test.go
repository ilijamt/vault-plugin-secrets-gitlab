//go:build paths || saas || selfhosted || e2e

package integration_test

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	t "github.com/ilijamt/vault-plugin-secrets-gitlab/internal/token"
	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/utils"
)

var (
	gitlabComPersonalAccessToken = cmp.Or(os.Getenv("GITLAB_COM_TOKEN"), "glpat-invalid-value")
	gitlabComUrl                 = cmp.Or(os.Getenv("GITLAB_COM_URL"), "https://gitlab.com")
	gitlabServiceAccountUrl      = cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_URL"), "http://localhost:8080")
	gitlabServiceAccountToken    = cmp.Or(os.Getenv("GITLAB_SERVICE_ACCOUNT_TOKEN"), "REPLACED-TOKEN")
)

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

func parseTimeFromFile(name string) (t time.Time, err error) {
	var buff []byte
	if buff, err = os.ReadFile(fmt.Sprintf("./testdata/%s", name)); err != nil {
		return t, err
	}
	return time.Parse(time.RFC3339, string(buff))
}

func ctxTestTime(ctx context.Context, tb testing.TB, tokenName string) (_ context.Context, t time.Time) {
	tb.Helper()
	var token = getGitlabToken(tokenName)
	if token.Empty() {
		var err error
		switch tb.Name() {
		case "TestGitlabClient_InvalidToken":
			// no token for this test
		case "TestWithGitlabUser_RotateToken":
			if t, err = parseTimeFromFile("gitlab-com"); err != nil {
				panic(err)
			}
		case "TestWithServiceAccountUser",
			"TestWithServiceAccountGroup",
			"TestWithServiceAccountUserFail_dedicated",
			"TestWithServiceAccountUserFail_saas":
			if t, err = parseTimeFromFile("gitlab-selfhosted"); err != nil {
				panic(err)
			}
		default:
			tb.Fatalf("unknown test name %s", tb.Name())
		}
	} else {
		t = token.CreatedAtTime()
	}
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
	ID        string `json:"id"`
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

func (g generatedToken) Empty() bool {
	return generatedToken{} == g
}

const (
	gitlabTimeLayout = "2006-01-02 15:04:05.000 -0700 MST"
)

func (g generatedToken) CreatedAtTime() (t time.Time) {
	t, _ = time.Parse(gitlabTimeLayout, g.CreatedAt)
	return t
}

type generatedTokens map[string]generatedToken

var loadTokens = sync.OnceValues(func() (t generatedTokens, err error) {
	version := os.Getenv("GITLAB_VERSION")
	if version == "" {
		return t, errors.New("GITLAB_VERSION env var must be set to load tokens; run via 'make test' or export GITLAB_VERSION explicitly")
	}
	var payload []byte
	if payload, err = os.ReadFile(fmt.Sprintf("./testdata/tokens.%s.json", version)); err != nil {
		return t, err
	}

	err = json.Unmarshal(payload, &t)
	return t, err
})

func getGitlabToken(name string) generatedToken {
	var tokens, _ = loadTokens()
	if token, ok := tokens[name]; ok {
		return token
	}
	return generatedToken{}
}
