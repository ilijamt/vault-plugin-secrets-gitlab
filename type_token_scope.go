package gitlab

import (
	"errors"
	"fmt"
	"slices"
)

type TokenScope string

const (
	// TokenScopeApi grants complete read/write access to the API, including all groups and projects, the container registry, the dependency proxy, and the package registry. Also grants complete read/write access to the registry and repository using Git over HTTP
	TokenScopeApi = TokenScope("api")
	// TokenScopeReadApi grants read access to the scoped group and related project API, including the Package Registry
	TokenScopeReadApi = TokenScope("read_api")
	// TokenScopeReadRegistry grants read access (pull) to the Container Registry images if any project within expected group is private and authorization is required.
	TokenScopeReadRegistry = TokenScope("read_registry")
	// TokenScopeWriteRegistry grants write access (push) to the Container Registry.
	TokenScopeWriteRegistry = TokenScope("write_registry")
	// TokenScopeReadRepository grants read access (pull) to the Container Registry images if any project within expected group is private and authorization is required
	TokenScopeReadRepository = TokenScope("read_repository")
	// TokenScopeWriteRepository grants read and write access (pull and push) to all repositories within expected group
	TokenScopeWriteRepository = TokenScope("write_repository")

	// TokenScopeReadPackageRegistry Allows read-only access to the package registry.
	TokenScopeReadPackageRegistry = TokenScope("read_package_registry")
	// TokenScopeWritePackageRegistry Allows read and write access to the package registry.
	TokenScopeWritePackageRegistry = TokenScope("write_package_registry")

	// TokenScopeCreateRunner grants permission to create runners in expected group
	TokenScopeCreateRunner = TokenScope("create_runner")
	// TokenScopeManageRunner grants permission to manage runners in expected group
	TokenScopeManageRunner = TokenScope("manage_runner")

	// TokenScopeReadUser grants read-only access to the authenticated user’s profile through the /user API endpoint, which includes username, public email, and full name. Also grants access to read-only API endpoints under /users.
	TokenScopeReadUser = TokenScope("read_user")
	// TokenScopeSudo grants permission to perform API actions as any user in the system, when authenticated as an administrator.
	TokenScopeSudo = TokenScope("sudo")
	// TokenScopeAdminMode grants permission to perform API actions as an administrator, when Admin Mode is enabled.
	TokenScopeAdminMode = TokenScope("admin_mode")

	// TokenScopeAiFeatures grants permission to perform API actions for GitLab Duo. This scope is designed to work with the GitLab Duo Plugin for JetBrains. For all other extensions, see scope requirements.
	TokenScopeAiFeatures = TokenScope("ai_features")
	// TokenScopeK8SProxy grants permission to perform Kubernetes API calls using the agent for Kubernetes.
	TokenScopeK8SProxy = TokenScope("k8s_proxy")
	// TokenScopeReadServicePing grant access to download Service Ping payload through the API when authenticated as an admin use.
	TokenScopeReadServicePing = TokenScope("read_service_ping")

	// TokenScopeSelfRotate grants permission to rotate this token using the personal access token API. Does not allow rotation of other tokens.
	TokenScopeSelfRotate = TokenScope("self_rotate")
	// TokenScopeReadVirtualRegistry if a project is private and authorization is required, grants read-only (pull) access to container images through the dependency proxy. Available only when the dependency proxy is enabled.
	TokenScopeReadVirtualRegistry = TokenScope("read_virtual_registry")
	// TokenScopeWriteVirtualRegistry if a project is private and authorization is required, grants read (pull), write (push), and delete access to container images through the dependency proxy. Available only when the dependency proxy is enabled.
	TokenScopeWriteVirtualRegistry = TokenScope("write_virtual_registry")

	TokenScopeUnknown = TokenScope("")
)

var (
	ErrUnknownTokenScope = errors.New("unknown token scope")

	// ValidPersonalTokenScopes defines the actions you can perform when you authenticate with a project access token.
	ValidPersonalTokenScopes = []string{
		TokenScopeApi.String(),
		TokenScopeReadUser.String(),
		TokenScopeReadApi.String(),
		TokenScopeReadRepository.String(),
		TokenScopeWriteRepository.String(),
		TokenScopeReadRegistry.String(),
		TokenScopeWriteRegistry.String(),
		TokenScopeReadVirtualRegistry.String(),
		TokenScopeWriteVirtualRegistry.String(),
		TokenScopeSudo.String(),
		TokenScopeAdminMode.String(),
		TokenScopeCreateRunner.String(),
		TokenScopeManageRunner.String(),
		TokenScopeAiFeatures.String(),
		TokenScopeK8SProxy.String(),
		TokenScopeSelfRotate.String(),
		TokenScopeReadServicePing.String(),
	}

	ValidProjectTokenScopes = []string{
		TokenScopeApi.String(),
		TokenScopeReadApi.String(),
		TokenScopeReadRegistry.String(),
		TokenScopeWriteRegistry.String(),
		TokenScopeReadRepository.String(),
		TokenScopeWriteRepository.String(),
		TokenScopeCreateRunner.String(),
		TokenScopeManageRunner.String(),
		TokenScopeAiFeatures.String(),
		TokenScopeK8SProxy.String(),
		TokenScopeSelfRotate.String(),
	}

	ValidGroupTokenScopes = []string{
		TokenScopeApi.String(),
		TokenScopeReadApi.String(),
		TokenScopeReadRegistry.String(),
		TokenScopeWriteRegistry.String(),
		TokenScopeReadVirtualRegistry.String(),
		TokenScopeWriteVirtualRegistry.String(),
		TokenScopeReadRepository.String(),
		TokenScopeWriteRepository.String(),
		TokenScopeCreateRunner.String(),
		TokenScopeManageRunner.String(),
		TokenScopeAiFeatures.String(),
		TokenScopeK8SProxy.String(),
		TokenScopeSelfRotate.String(),
	}

	ValidUserServiceAccountTokenScopes = ValidPersonalTokenScopes

	ValidGroupServiceAccountTokenScopes = ValidGroupTokenScopes

	ValidPipelineProjectTokenScopes []string

	ValidProjectDeployTokenScopes = []string{
		TokenScopeReadRepository.String(),
		TokenScopeReadRegistry.String(),
		TokenScopeWriteRegistry.String(),
		TokenScopeReadVirtualRegistry.String(),
		TokenScopeWriteVirtualRegistry.String(),
		TokenScopeReadPackageRegistry.String(),
		TokenScopeWritePackageRegistry.String(),
	}

	ValidGroupDeployTokenScopes = []string{
		TokenScopeReadRepository.String(),
		TokenScopeReadRegistry.String(),
		TokenScopeWriteRegistry.String(),
		TokenScopeReadVirtualRegistry.String(),
		TokenScopeWriteVirtualRegistry.String(),
		TokenScopeReadPackageRegistry.String(),
		TokenScopeWritePackageRegistry.String(),
	}
)

func (i TokenScope) String() string {
	return string(i)
}

func (i TokenScope) Value() string {
	return i.String()
}

func TokenScopeParse(value string) (TokenScope, error) {
	if slices.Contains(ValidGroupTokenScopes, value) ||
		slices.Contains(ValidPipelineProjectTokenScopes, value) ||
		slices.Contains(ValidGroupDeployTokenScopes, value) ||
		slices.Contains(ValidProjectDeployTokenScopes, value) ||
		slices.Contains(ValidPersonalTokenScopes, value) ||
		slices.Contains(ValidProjectTokenScopes, value) ||
		slices.Contains(ValidUserServiceAccountTokenScopes, value) ||
		slices.Contains(ValidGroupServiceAccountTokenScopes, value) {
		return TokenScope(value), nil
	}
	return TokenScopeUnknown, fmt.Errorf("failed to parse '%s': %w", value, ErrUnknownTokenScope)
}
