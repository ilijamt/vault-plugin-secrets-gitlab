//go:build g16x

package token

var (
	ValidPersonalScopes = []string{
		ScopeApi.String(),
		ScopeReadUser.String(),
		ScopeReadApi.String(),
		ScopeReadRepository.String(),
		ScopeWriteRepository.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeSudo.String(),
		ScopeAdminMode.String(),
		ScopeCreateRunner.String(),
		ScopeAiFeatures.String(),
		ScopeK8SProxy.String(),
		ScopeReadServicePing.String(),
	}

	ValidProjectScopes = []string{
		ScopeApi.String(),
		ScopeReadApi.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadRepository.String(),
		ScopeWriteRepository.String(),
		ScopeCreateRunner.String(),
		ScopeAiFeatures.String(),
		ScopeK8SProxy.String(),
	}

	ValidGroupScopes = []string{
		ScopeApi.String(),
		ScopeReadApi.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadRepository.String(),
		ScopeWriteRepository.String(),
		ScopeCreateRunner.String(),
		ScopeAiFeatures.String(),
		ScopeK8SProxy.String(),
	}

	ValidUserServiceAccountScopes = ValidPersonalScopes

	ValidGroupServiceAccountScopes = ValidGroupScopes

	ValidPipelineProjectScopes []string

	ValidProjectDeployScopes = []string{
		ScopeReadRepository.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadPackageRegistry.String(),
		ScopeWritePackageRegistry.String(),
	}

	ValidGroupDeployScopes = []string{
		ScopeReadRepository.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadPackageRegistry.String(),
		ScopeWritePackageRegistry.String(),
	}
)
