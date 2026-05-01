package token

import (
	"fmt"
	"sort"

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

// scopeMinVersionByTokenType maps a token type to the scopes it supports and
// the GitLab MAJOR.MINOR version each became available.
//
// A token type whose value is nil indicates "scopes is not applicable for
// this token type" (pipeline trigger). "0.0" means always allowed within the
// supported window. Scopes whose introduction version is unverified
// (create_runner, ai_features) are recorded as "0.0" — the lenient default
// per the version-aware-tokens plan.
var scopeMinVersionByTokenType = map[Type]map[Scope]string{
	TypePersonal: {
		ScopeApi:                  "0.0",
		ScopeReadApi:              "0.0",
		ScopeReadUser:             "0.0",
		ScopeReadRepository:       "0.0",
		ScopeWriteRepository:      "0.0",
		ScopeReadRegistry:         "0.0",
		ScopeWriteRegistry:        "0.0",
		ScopeReadVirtualRegistry:  "18.0",
		ScopeWriteVirtualRegistry: "18.0",
		ScopeSudo:                 "0.0",
		ScopeAdminMode:            "0.0",
		ScopeCreateRunner:         "0.0",
		ScopeManageRunner:         "17.1",
		ScopeAiFeatures:           "0.0",
		ScopeK8SProxy:             "0.0",
		ScopeSelfRotate:           "17.9",
		ScopeReadServicePing:      "17.1",
	},
	TypeProject: {
		ScopeApi:             "0.0",
		ScopeReadApi:         "0.0",
		ScopeReadRegistry:    "0.0",
		ScopeWriteRegistry:   "0.0",
		ScopeReadRepository:  "0.0",
		ScopeWriteRepository: "0.0",
		ScopeCreateRunner:    "0.0",
		ScopeManageRunner:    "17.1",
		ScopeAiFeatures:      "0.0",
		ScopeK8SProxy:        "0.0",
		ScopeSelfRotate:      "17.9",
	},
	TypeGroup: {
		ScopeApi:                  "0.0",
		ScopeReadApi:              "0.0",
		ScopeReadRegistry:         "0.0",
		ScopeWriteRegistry:        "0.0",
		ScopeReadVirtualRegistry:  "18.0",
		ScopeWriteVirtualRegistry: "18.0",
		ScopeReadRepository:       "0.0",
		ScopeWriteRepository:      "0.0",
		ScopeCreateRunner:         "0.0",
		ScopeManageRunner:         "17.1",
		ScopeAiFeatures:           "0.0",
		ScopeK8SProxy:             "0.0",
		ScopeSelfRotate:           "17.9",
	},
	TypeProjectDeploy: {
		ScopeReadRepository:       "0.0",
		ScopeReadRegistry:         "0.0",
		ScopeWriteRegistry:        "0.0",
		ScopeReadVirtualRegistry:  "18.0",
		ScopeWriteVirtualRegistry: "18.0",
		ScopeReadPackageRegistry:  "0.0",
		ScopeWritePackageRegistry: "0.0",
	},
	TypeGroupDeploy: {
		ScopeReadRepository:       "0.0",
		ScopeReadRegistry:         "0.0",
		ScopeWriteRegistry:        "0.0",
		ScopeReadVirtualRegistry:  "18.0",
		ScopeWriteVirtualRegistry: "18.0",
		ScopeReadPackageRegistry:  "0.0",
		ScopeWritePackageRegistry: "0.0",
	},
	TypeUserServiceAccount: {
		ScopeApi:                  "0.0",
		ScopeReadApi:              "0.0",
		ScopeReadUser:             "0.0",
		ScopeReadRepository:       "0.0",
		ScopeWriteRepository:      "0.0",
		ScopeReadRegistry:         "0.0",
		ScopeWriteRegistry:        "0.0",
		ScopeReadVirtualRegistry:  "18.0",
		ScopeWriteVirtualRegistry: "18.0",
		ScopeSudo:                 "0.0",
		ScopeAdminMode:            "0.0",
		ScopeCreateRunner:         "0.0",
		ScopeManageRunner:         "17.1",
		ScopeAiFeatures:           "0.0",
		ScopeK8SProxy:             "0.0",
		ScopeSelfRotate:           "17.9",
		ScopeReadServicePing:      "17.1",
	},
	TypeGroupServiceAccount: {
		ScopeApi:                  "0.0",
		ScopeReadApi:              "0.0",
		ScopeReadRegistry:         "0.0",
		ScopeWriteRegistry:        "0.0",
		ScopeReadVirtualRegistry:  "18.0",
		ScopeWriteVirtualRegistry: "18.0",
		ScopeReadRepository:       "0.0",
		ScopeWriteRepository:      "0.0",
		ScopeCreateRunner:         "0.0",
		ScopeManageRunner:         "17.1",
		ScopeAiFeatures:           "0.0",
		ScopeK8SProxy:             "0.0",
		ScopeSelfRotate:           "17.9",
	},
	TypePipelineProjectTrigger: nil, // not applicable
}

// ValidScopesFor returns the scopes allowed for tokenType on the given GitLab
// version, sorted alphabetically. applicable is false if tokenType does not
// take a scopes field (pipeline trigger). When version is empty the gate is
// lenient — every scope the token type accepts is returned.
func ValidScopesFor(tokenType Type, gitlabVersion string) (scopes []Scope, applicable bool) {
	inner, present := scopeMinVersionByTokenType[tokenType]
	if !present || inner == nil {
		return nil, false
	}
	for s, minV := range inner {
		if atLeast(gitlabVersion, minV) {
			scopes = append(scopes, s)
		}
	}
	sort.Slice(scopes, func(i, j int) bool { return scopes[i] < scopes[j] })
	return scopes, true
}

// IsScopeAllowed reports whether scope is a valid scope for tokenType on
// gitlabVersion. Returns false if tokenType does not take a scopes field.
func IsScopeAllowed(tokenType Type, scope Scope, gitlabVersion string) bool {
	inner, present := scopeMinVersionByTokenType[tokenType]
	if !present || inner == nil {
		return false
	}
	minV, ok := inner[scope]
	if !ok {
		return false
	}
	return atLeast(gitlabVersion, minV)
}

// AllValidScopes returns the union of scopes accepted by any token type at
// any version — used to populate the OpenAPI schema's AllowedValues at
// backend startup, before a GitLab version is known.
func AllValidScopes() []string {
	seen := map[Scope]struct{}{}
	for _, inner := range scopeMinVersionByTokenType {
		for s := range inner {
			seen[s] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s.String())
	}
	sort.Strings(out)
	return out
}

func (i Scope) String() string {
	return string(i)
}

func ParseScope(value string) (Scope, error) {
	for _, inner := range scopeMinVersionByTokenType {
		if _, ok := inner[Scope(value)]; ok {
			return Scope(value), nil
		}
	}
	return ScopeUnknown, fmt.Errorf("failed to parse '%s': %w", value, errs.ErrUnknownTokenScope)
}
