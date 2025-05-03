package goreleaser

import "list"

// Generate the builds array by combining the base config with each GitLab version
builds: [
	for version, config in gitlabVersions {
		baseBuildConfig & {
			if config.extraLdflags != [] {
				ldflags: list.Concat([baseBuildConfig.ldflags, config.extraLdflags])
			}
			if config.suffix != "" {
				binary: "{{ .ProjectName }}_\(config.suffix)v{{ .Version }}"
			}
		}
	},
]

// Output the final goreleaser configuration
output: base & {
	builds:   builds
	archives: archiveConfig
}
