//go:build unit || saas || selfhosted || local

package gitlab_test

import (
	"cmp"
	"fmt"
	"net/http"
	"os"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func getClient(t *testing.T, target string) (client *http.Client, u string) {
	t.Helper()

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

	u = cmp.Or(os.Getenv("GITLAB_URL"), "http://localhost:8080/")
	return r.GetDefaultClient(), u
}
