//go:build paths || saas || selfhosted || e2e

package integration_test

import (
	"bytes"
	"cmp"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

var tokenFieldRegex = regexp.MustCompile(`"token":"[^"]*"`)

type skipEmptyFS struct {
	cassette.FS
}

func (c *skipEmptyFS) WriteFile(name string, data []byte) error {
	if bytes.Contains(data, []byte("interactions: []")) {
		return nil
	}
	return c.FS.WriteFile(name, data)
}

// getClient returns an HTTP client backed by a go-vcr recorder for the given
// target. Cassettes live under tests/integration/testdata/<target>/...
//
//   - "paths" and "e2e" require GITLAB_VERSION because cassettes are recorded
//     against a specific self-hosted GitLab version. Layout:
//     testdata/paths/<version>/<TestName>.yaml
//     testdata/e2e/<version>/<TestName>.yaml
//
//   - "saas" and "selfhosted" do not pin a version: cassettes target gitlab.com
//     or a long-lived self-hosted instance and remain valid across patch
//     versions of GitLab. Layout:
//     testdata/saas/<TestName>.yaml
//     testdata/selfhosted/<TestName>.yaml
//
// `make test` sets GITLAB_VERSION; running paths/e2e tests directly requires
// exporting it. Test names are filename-sanitized via sanitizePath.
func getClient(t *testing.T, target string) (client *http.Client, u string) {
	t.Helper()

	defaultGitlabHost := "localhost:8080"

	var filename string
	switch target {
	case "paths", "e2e":
		version := os.Getenv("GITLAB_VERSION")
		if version == "" {
			t.Fatal("GITLAB_VERSION env var must be set for paths/e2e cassettes; run via 'make test' or export GITLAB_VERSION explicitly")
		}
		filename = fmt.Sprintf("testdata/%s/%s/%s", target, version, sanitizePath(t.Name()))
	default:
		filename = fmt.Sprintf("testdata/%s/%s", target, sanitizePath(t.Name()))
	}
	r, err := recorder.New(filename,
		[]recorder.Option{
			recorder.WithSkipRequestLatency(os.Getenv("SKIP_REQUEST_LATENCY") != ""),
			recorder.WithMode(recorder.ModeRecordOnce),
			recorder.WithFS(&skipEmptyFS{FS: cassette.NewDiskFS()}),
			recorder.WithMatcher(
				cassette.NewDefaultMatcher(
					cassette.WithIgnoreUserAgent(),
					cassette.WithIgnoreAuthorization(),
					cassette.WithIgnoreHeaders(
						"Private-Token",
					),
				),
			),
			recorder.WithHook(func(i *cassette.Interaction) error {
				i.Request.Headers.Set("Private-Token", "REPLACED-TOKEN")
				i.Response.Body = tokenFieldRegex.ReplaceAllString(i.Response.Body, `"token":"REPLACED-TOKEN"`)
				return nil
			}, recorder.BeforeSaveHook),
		}...,
	)
	if err != nil {
		t.Fatalf("could not create recorder: %s", err)
	}

	t.Cleanup(func() {
		if err := r.Stop(); err != nil {
			t.Errorf("could not close recorder: %s", err)
		}
	})

	u = cmp.Or(os.Getenv("GITLAB_URL"), fmt.Sprintf("http://%s/", defaultGitlabHost))
	return r.GetDefaultClient(), u
}
