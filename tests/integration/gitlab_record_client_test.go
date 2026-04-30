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

	filename := fmt.Sprintf("testdata/%s/%s", target, sanitizePath(t.Name()))
	r, err := recorder.New(filename,
		[]recorder.Option{
			recorder.WithSkipRequestLatency(false),
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
				// u, _ := url.Parse(i.Request.URL)
				// u.Host = defaultGitlabHost
				// i.Request.Host = u.Host
				// i.Request.URL = u.String()
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
