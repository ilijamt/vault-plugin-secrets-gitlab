//go:build !g16x

package token

var (
	// ValidPersonalScopes defines the actions you can perform when you authenticate with a project access token.
	ValidPersonalScopes = []string{
		ScopeApi.String(),
		ScopeReadUser.String(),
		ScopeReadApi.String(),
		ScopeReadRepository.String(),
		ScopeWriteRepository.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadVirtualRegistry.String(),
		ScopeWriteVirtualRegistry.String(),
		ScopeSudo.String(),
		ScopeAdminMode.String(),
		ScopeCreateRunner.String(),
		ScopeManageRunner.String(),
		ScopeAiFeatures.String(),
		ScopeK8SProxy.String(),
		ScopeSelfRotate.String(),
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
		ScopeManageRunner.String(),
		ScopeAiFeatures.String(),
		ScopeK8SProxy.String(),
		ScopeSelfRotate.String(),
	}

	ValidGroupScopes = []string{
		ScopeApi.String(),
		ScopeReadApi.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadVirtualRegistry.String(),
		ScopeWriteVirtualRegistry.String(),
		ScopeReadRepository.String(),
		ScopeWriteRepository.String(),
		ScopeCreateRunner.String(),
		ScopeManageRunner.String(),
		ScopeAiFeatures.String(),
		ScopeK8SProxy.String(),
		ScopeSelfRotate.String(),
	}

	ValidUserServiceAccountScopes = ValidPersonalScopes

	ValidGroupServiceAccountScopes = ValidGroupScopes

	ValidPipelineProjectScopes []string

	ValidProjectDeployScopes = []string{
		ScopeReadRepository.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadVirtualRegistry.String(),
		ScopeWriteVirtualRegistry.String(),
		ScopeReadPackageRegistry.String(),
		ScopeWritePackageRegistry.String(),
	}

	ValidGroupDeployScopes = []string{
		ScopeReadRepository.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadVirtualRegistry.String(),
		ScopeWriteVirtualRegistry.String(),
		ScopeReadPackageRegistry.String(),
		ScopeWritePackageRegistry.String(),
	}
)
