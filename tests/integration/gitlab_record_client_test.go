//go:build paths || saas || serviceaccount || e2e

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
// target, with the cassette path resolved by cassettePath.
func getClient(t *testing.T, target string) (client *http.Client, u string) {
	t.Helper()

	defaultGitlabHost := "localhost:8080"

	r, err := recorder.New(cassettePath(t, target),
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
