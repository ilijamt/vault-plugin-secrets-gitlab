//go:build unit || saas || selfhosted || local

package integration_test

import (
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

func getClient(t *testing.T, target string) (client *http.Client, u string) {
	t.Helper()

	defaultGitlabHost := "localhost:8080"

	var filename string
	switch target {
	case "unit", "local":
		version := os.Getenv("GITLAB_VERSION")
		if version == "" {
			t.Fatal("GITLAB_VERSION env var must be set for unit/local cassettes; run via 'make test' or export GITLAB_VERSION explicitly")
		}
		filename = fmt.Sprintf("testdata/%s/%s/%s", target, version, sanitizePath(t.Name()))
	default:
		filename = fmt.Sprintf("testdata/%s/%s", target, sanitizePath(t.Name()))
	}
	r, err := recorder.New(filename,
		[]recorder.Option{
			recorder.WithSkipRequestLatency(os.Getenv("SKIP_REQUEST_LATENCY") != ""),
			recorder.WithMode(recorder.ModeRecordOnce),
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
