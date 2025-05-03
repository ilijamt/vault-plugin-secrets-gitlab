package goreleaser

// Define the common build configuration
baseBuildConfig: {
	env: ["CGO_ENABLED=0"]
	main:          "./cmd/vault-plugin-secrets-gitlab/main.go"
	mod_timestamp: "{{ .CommitTimestamp }}"
	flags: ["-trimpath"]
	ldflags: [
		"-s -w",
		"-X 'github.com/ilijamt/vault-plugin-secrets-gitlab.Version=v{{ .Version }}'",
		"-X 'github.com/ilijamt/vault-plugin-secrets-gitlab.FullCommit={{ .FullCommit }}'",
		"-X 'github.com/ilijamt/vault-plugin-secrets-gitlab.BuildDate={{ .Date }}'",
	]
	goos: [
		"windows",
		"linux",
		"darwin",
		"illumos",
	]
	goarch: [
		"amd64",
		"386",
		"arm",
		"arm64",
	]
	ignore: [
		{
			goos:   "darwin"
			goarch: "386"
		},
	]
	binary: "{{ .ProjectName }}_v{{ .Version }}"
}
