package gitlab_test

import (
	"cmp"
	"fmt"
	"net/http"
	"os"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

func getClient(t *testing.T) (client *http.Client, u string) {
	t.Helper()

	filename := fmt.Sprintf("testdata/fixtures/%s/%s", gitlabVersion, sanitizePath(t.Name()))
	r, err := recorder.New(filename)
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
