package goreleaser

// Base configuration that will be shared across all variants
base: {
	version:      "2"
	report_sizes: true
	sboms: [
		{artifacts: "archive"},
	]
	checksum: {
		name_template: "{{ .ProjectName }}_{{ .Version }}_SHA256SUMS"
		algorithm:     "sha256"
	}
	changelog: {
		sort: "asc"
		use:  "github"
		filters: {
			exclude: [
				"^docs:",
				"^test:",
				"merge conflict",
				"Merge pull request",
				"Merge remote-tracking branch",
				"Merge branch",
				"go mod tidy",
			]
		}
	}
}

// Define the archive configuration
archiveConfig: [
	{
		formats: ["tar.gz"]
		name_template: """
			{{ .ProjectName }}_
			{{- .Os }}_
			{{- if eq .Arch "amd64" }}x86_64
			{{- else if eq .Arch "386" }}i386
			{{- else }}{{ .Arch }}{{ end }}
			{{- if .Arm }}v{{ .Arm }}{{ end }}
			"""
		format_overrides: [
			{
				goos: "windows"
				formats: ["zip"]
			},
		]
	},
]
