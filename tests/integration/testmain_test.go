//go:build paths || saas || selfhosted || e2e

package integration_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

var (
	tagPaths bool
	tagE2E   bool
)

func TestMain(m *testing.M) {
	if msg := validateIntegrationEnv(); msg != "" {
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func validateIntegrationEnv() string {
	var missing []string
	if (tagPaths || tagE2E) && os.Getenv("GITLAB_VERSION") == "" {
		missing = append(missing, "GITLAB_VERSION")
	}
	if len(missing) == 0 {
		return ""
	}
	return fmt.Sprintf(
		"integration tests missing required env vars: %s\n"+
			"Run via 'make test' or export them explicitly.",
		strings.Join(missing, ", "),
	)
}
