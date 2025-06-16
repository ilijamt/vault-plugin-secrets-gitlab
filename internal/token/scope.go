package token

import (
	"fmt"
	"slices"

	"github.com/ilijamt/vault-plugin-secrets-gitlab/internal/errs"
)

type Scope string

const (
	// ScopeApi grants complete read/write access to the API, including all groups and projects, the container registry, the dependency proxy, and the package registry. Also grants complete read/write access to the registry and repository using Git over HTTP
	ScopeApi = Scope("api")
	// ScopeReadApi grants read access to the scoped group and related project API, including the Package Registry
	ScopeReadApi = Scope("read_api")
	// ScopeReadRegistry grants read access (pull) to the Container Registry images if any project within expected group is private and authorization is required.
	ScopeReadRegistry = Scope("read_registry")
	// ScopeWriteRegistry grants write access (push) to the Container Registry.
	ScopeWriteRegistry = Scope("write_registry")
	// ScopeReadRepository grants read access (pull) to the Container Registry images if any project within expected group is private and authorization is required
	ScopeReadRepository = Scope("read_repository")
	// ScopeWriteRepository grants read and write access (pull and push) to all repositories within expected group
	ScopeWriteRepository = Scope("write_repository")

	// ScopeReadPackageRegistry Allows read-only access to the package registry.
	ScopeReadPackageRegistry = Scope("read_package_registry")
	// ScopeWritePackageRegistry Allows read and write access to the package registry.
	ScopeWritePackageRegistry = Scope("write_package_registry")

	// ScopeCreateRunner grants permission to create runners in expected group
	ScopeCreateRunner = Scope("create_runner")
	// ScopeManageRunner grants permission to manage runners in expected group
	ScopeManageRunner = Scope("manage_runner")

	// ScopeReadUser grants read-only access to the authenticated user’s profile through the /user API endpoint, which includes username, public email, and full name. Also grants access to read-only API endpoints under /users.
	ScopeReadUser = Scope("read_user")
	// ScopeSudo grants permission to perform API actions as any user in the system, when authenticated as an administrator.
	ScopeSudo = Scope("sudo")
	// ScopeAdminMode grants permission to perform API actions as an administrator, when Admin Mode is enabled.
	ScopeAdminMode = Scope("admin_mode")

	// ScopeAiFeatures grants permission to perform API actions for GitLab Duo. This scope is designed to work with the GitLab Duo Plugin for JetBrains. For all other extensions, see scope requirements.
	ScopeAiFeatures = Scope("ai_features")
	// ScopeK8SProxy grants permission to perform Kubernetes API calls using the agent for Kubernetes.
	ScopeK8SProxy = Scope("k8s_proxy")
	// ScopeReadServicePing grant access to download Service Ping payload through the API when authenticated as an admin use.
	ScopeReadServicePing = Scope("read_service_ping")

	// ScopeSelfRotate grants permission to rotate this token using the personal access token API. Does not allow rotation of other tokens.
	ScopeSelfRotate = Scope("self_rotate")
	// ScopeReadVirtualRegistry if a project is private and authorization is required, grants read-only (pull) access to container images through the dependency proxy. Available only when the dependency proxy is enabled.
	ScopeReadVirtualRegistry = Scope("read_virtual_registry")
	// ScopeWriteVirtualRegistry if a project is private and authorization is required, grants read (pull), write (push), and delete access to container images through the dependency proxy. Available only when the dependency proxy is enabled.
	ScopeWriteVirtualRegistry = Scope("write_virtual_registry")

	ScopeUnknown = Scope("")
)

var (

	// ValidPersonalTokenScopes defines the actions you can perform when you authenticate with a project access token.
	ValidPersonalTokenScopes = []string{
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

	ValidProjectTokenScopes = []string{
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

	ValidGroupTokenScopes = []string{
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

	ValidUserServiceAccountTokenScopes = ValidPersonalTokenScopes

	ValidGroupServiceAccountTokenScopes = ValidGroupTokenScopes

	ValidPipelineProjectTokenScopes []string

	ValidProjectDeployTokenScopes = []string{
		ScopeReadRepository.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadVirtualRegistry.String(),
		ScopeWriteVirtualRegistry.String(),
		ScopeReadPackageRegistry.String(),
		ScopeWritePackageRegistry.String(),
	}

	ValidGroupDeployTokenScopes = []string{
		ScopeReadRepository.String(),
		ScopeReadRegistry.String(),
		ScopeWriteRegistry.String(),
		ScopeReadVirtualRegistry.String(),
		ScopeWriteVirtualRegistry.String(),
		ScopeReadPackageRegistry.String(),
		ScopeWritePackageRegistry.String(),
	}
)

func (i Scope) String() string {
	return string(i)
}

func (i Scope) Value() string {
	return i.String()
}

func ScopeParse(value string) (Scope, error) {
	if slices.Contains(ValidGroupTokenScopes, value) ||
		slices.Contains(ValidPipelineProjectTokenScopes, value) ||
		slices.Contains(ValidGroupDeployTokenScopes, value) ||
		slices.Contains(ValidProjectDeployTokenScopes, value) ||
		slices.Contains(ValidPersonalTokenScopes, value) ||
		slices.Contains(ValidProjectTokenScopes, value) ||
		slices.Contains(ValidUserServiceAccountTokenScopes, value) ||
		slices.Contains(ValidGroupServiceAccountTokenScopes, value) {
		return Scope(value), nil
	}
	return ScopeUnknown, fmt.Errorf("failed to parse '%s': %w", value, errs.ErrUnknownTokenScope)
}
