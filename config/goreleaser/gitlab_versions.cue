package goreleaser

// Generate GitLab version-specific build configurations
gitlabVersions: {
	default: {
		suffix: "g17x_"
		extraLdflags: [
			"-X 'github.com/ilijamt/vault-plugin-secrets-gitlab.VersionGitlab=17'",
		]
	}

	// GitLab 16 build
	gitlab16: {
		suffix: "g16x_"
		extraLdflags: [
			"-X 'github.com/ilijamt/vault-plugin-secrets-gitlab.VersionGitlab=16'",
		]
	}
}
