# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]


### <!-- 0 -->Features

- *(local-env)* Added input validation for GitLab version in setup scripts

### <!-- 4 -->Documentation

- Clarify plugin support for OpenBao
- *(changelog)* Regenerate CHANGELOG.md [skip ci]
- *(changelog)* Regenerate CHANGELOG.md [skip ci]
- *(changelog)* Regenerate CHANGELOG.md [skip ci]
- *(changelog)* Regenerate CHANGELOG.md [skip ci]
- *(changelog)* Regenerate CHANGELOG.md [skip ci]
- *(changelog)* Regenerate CHANGELOG.md [skip ci]
- Updated regex pattern in endpoint configuration docs

### <!-- 5 -->Tests

- Upgrade to 17.11.7 for local and unit testing

### <!-- 6 -->CI

- *(workflows)* Configured changelog generation for automated releases
- *(workflows)* Configured matrix testing for GitLab version discovery
- *(workflows)* Configured artifact path to exclude version directories

### <!-- 7 -->Dependencies

- *(deps)* Bump actions/checkout from 4 to 6 (#331) (#331)
- *(deps)* Bump softprops/action-gh-release from 2 to 3 (#330)

### <!-- 9 -->Chore

- Configured git-cliff for changelog generation
- *(workflows)* Updated changelog workflow to use git status
- Parameterized GitLab version in local setup scripts
- *(workflows)* Updated GitHub Actions for artifact actions

## [0.11.0] - 2026-04-20


### <!-- 0 -->Features

- Introduced secret package for GitLab token management
- Introduced config and role methods to Backend for better management
- Introduced modular backend interfaces and GitLab config paths
- Reordered parameters for consistency in SaveConfig and GetRole
- Introduced DeleteClient for improved client management
- *(local-env)* Introduced setup improvements with error control

### <!-- 1 -->Bug Fixes

- IsValidPath should return false if token type is not one of the handled cases
- *(backend)* Client did not get removed when called SetClient with nil
- *(paths/config)* Patched client locks with centralized mutex in config
- *(gitlab)* Corrected CreatePipelineProjectTriggerAccessToken logging
- *(model)* Resolved incorrect error messages for nil Storage refs
- *(model/token)* Dropped scopes in various tokens and simplified tests
- *(model/config)* Updated variable in entry_config for consistency
- Use GetGroup in GetGroupIdByPath for improved accuracy

### <!-- 3 -->Refactor

- Migrated to gitlab.Client for interface consistency
- *(backend)* Migrated interfaces to backend for modularity
- Restructured client handling in backend interfaces
- Renamed package to types for clarity
- *(backend)* Renamed BackendImpl to Impl for consistency
- *(backend)* Renamed ClientProvider to ClientReader in backend
- Restructured tests with struct-based stubs, removed objx
- *(paths/config)* Migrated key parsing to strings.Cut for clarity
- Replaced gitlab with backend in tests and removed defs.go
- *(paths/config)* Restructured variable init for config retrieval
- Migrated locking interfaces to unify locking methods
- *(model)* Removed unused Event interface from model code
- *(token)* Renamed AccessLevelParse to ParseAccessLevel for naming consistency
- *(token)* Centralized error handling using errs package
- *(token)* Removed redundant methods and assertions from token

### <!-- 4 -->Documentation

- Split and reorganize the docs
- *(k8s-external-secrets-operator)* Documented Vault integration
- *(token)* Corrected comment on personal access token usage

### <!-- 5 -->Tests

- Added some edge case tests
- Fix test for gitlab ctx
- *(mocks)* Introduced interfaces and tests to mock event and flag behavior
- *(secret)* Added HandleRevoke tests for error handling validation
- *(backend)* Verified replication state with TestPeriodicFunc_WriteSafeReplicationState
- Removed build constraints from test files
- *(paths/flags)* Expanded write tests with event checks and scenarios
- *(backend)* Introduced mock event sender for backend test validation
- *(paths/flags)* Asserted event type and metadata in TestPathFlagsUpdate
- *(paths/token)* Added tests for token role creation and validation
- *(paths/config)* Added tests for config path error handling
- *(event)* Substituted mockEventsSender with logical.MockEventSender
- *(paths/role)* Validated role path CRUD with mocks, utilities added
- *(backend)* Verified double-check locking in GetClientByName method
- *(gitlab)* Verified Gitlab client retrieval from context
- *(gitlab)* Introduced subtests for better granularity and clarity
- *(model/role)* Renamed test function from TestRule to TestRole
- *(token)* Introduced subtest names for improved isolation clarity
- *(token)* Introduced subtests for better granularity in token tests
- *(token)* Reorganized access level tests with subtests for clarity
- *(token)* Adapted test assertions to t.Run for isolated subtests
- *(utils)* Extracted current time logic from test loop for consistency
- *(integration)* Restructure test suite and regenerate local/unit testdata

### <!-- 6 -->CI

- *(workflows)* Added dependency review for pull requests
- *(workflows)* Introduced govulncheck for Go vulnerability scanning
- *(workflows)* Configured path filters and concurrency in CodeQL
- *(workflows)* Configured stale workflow with permissions and timeout
- *(workflows)* Configured concurrency to cancel in-progress runs
- *(workflows)* Configured path filters and concurrency for CI

### <!-- 7 -->Dependencies

- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#294) (#294)
- *(deps)* Bump gopkg.in/dnaeon/go-vcr.v4 from 4.0.3 to 4.0.6 (#264) (#264)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#295) (#295)
- *(deps)* Bump anchore/sbom-action from 0.22.1 to 0.22.2 (#297) (#297)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#299) (#299)
- *(deps)* Bump github.com/hashicorp/vault/sdk in the hashicorp group (#298) (#298)
- *(deps)* Bump github.com/hashicorp/vault/sdk in the hashicorp group (#300) (#300)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#301) (#301)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#302) (#302)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#304) (#304)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#306) (#306)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#308) (#308)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#309) (#309)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#311) (#311)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#316) (#316)
- *(deps)* Bump anchore/sbom-action from 0.22.2 to 0.24.0 (#324) (#324)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#321) (#321)
- *(deps)* Bump github.com/hashicorp/vault/sdk in the hashicorp group (#325) (#325)
- *(deps)* Bump goreleaser/goreleaser-action from 6.4.0 to 7.0.0 (#315) (#315)
- *(deps)* Bump golang.org/x/time from 0.14.0 to 0.15.0 (#323) (#323)
- *(deps)* Bump codecov/codecov-action from 5 to 6 (#328) (#328)
- *(deps)* Bump the hashicorp group across 1 directory with 2 updates (#327) (#327)
- *(deps)* Bump github.com/hashicorp/vault/sdk in the hashicorp group (#329) (#329)
- *(deps)* Removed unused dependency from go.mod
- *(deps)* Organized dependabot for weekly updates and testing groups
- *(deps)* Updated dependencies in go.mod for latest compatibility

### <!-- 8 -->Build

- Fix Makefile clean-coverage

### <!-- 9 -->Chore

- Removed unused file
- Excluded Terraform directories from indexing to boost performance
- Renamed target in Makefile for clarity
- Removed unused mock configurations from .mockery.yaml
- *(workflows)* Removed obsolete comments from workflows
- *(local-env)* Aligned test data path with integration tests

## [0.10.0] - 2026-01-29


### <!-- 0 -->Features

- Allow dynamic paths for roles (#288)

### <!-- 1 -->Bug Fixes

- Golangci-run lint issues in tests
- Added a check if backend is nil in the event handler
- Remove a non existent flag

### <!-- 3 -->Refactor

- Extract models and put them in their own pkg (#254)
- Remove prefix from event function and replace it with a constant

### <!-- 7 -->Dependencies

- *(deps)* Bump anchore/sbom-action from 0.20.8 to 0.20.9 (#252) (#252)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#253) (#253)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#256) (#256)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#258) (#258)
- *(deps)* Bump actions/checkout from 5 to 6 (#261) (#261)
- *(deps)* Bump golangci/golangci-lint-action from 8 to 9 (#257) (#257)
- *(deps)* Bump anchore/sbom-action from 0.20.9 to 0.20.10 (#259) (#259)
- *(deps)* Bump anchore/sbom-action from 0.20.10 to 0.20.11 (#272) (#272)
- *(deps)* Bump anchore/sbom-action from 0.20.11 to 0.22.0 (#286) (#286)
- *(deps)* Bump github.com/hashicorp/vault/sdk in the hashicorp group (#282) (#282)
- *(deps)* Bump google.golang.org/protobuf from 1.36.10 to 1.36.11 (#274) (#274)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go from 0.160.0 to 1.17.0 (#289) (#289)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#290) (#290)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#293) (#293)
- *(deps)* Bump anchore/sbom-action from 0.22.0 to 0.22.1 (#291) (#291)

### <!-- 8 -->Build

- Fix go test to not cache the result

### <!-- 9 -->Chore

- Fix empty space in flags
- Added Makefile for running tests
- Fixed Makefile to be able to run the plugin locally

## [0.9.0] - 2025-10-17


### <!-- 0 -->Features

- Added stringsSplit, trimSpace, stringsReplace to the template name functionality

### <!-- 1 -->Bug Fixes

- Handle unexpected missing token expiry by setting a default (#179)

### <!-- 3 -->Refactor

- Change the function name to remove Gitlab from the function
- Structure to make it easier to extend and modify functionality (#195)

### <!-- 4 -->Documentation

- Added release badge

### <!-- 7 -->Dependencies

- *(deps)* Bump anchore/sbom-action from 0.18.0 to 0.19.0 (#181) (#181)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#180) (#180)
- *(deps)* Bump anchore/sbom-action from 0.19.0 to 0.20.0 (#184)
- *(deps)* Bump golangci/golangci-lint-action from 7 to 8 (#183) (#183)
- *(deps)* Bump the hashicorp group across 1 directory with 2 updates (#188) (#188)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#186) (#186)
- *(deps)* Bump golang.org/x/time from 0.11.0 to 0.12.0 (#189) (#189)
- *(deps)* Bump github.com/hashicorp/vault/sdk in the hashicorp group (#190) (#190)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#193) (#193)
- *(deps)* Bump gopkg.in/dnaeon/go-vcr.v4 from 4.0.2 to 4.0.3 (#191) (#191)
- *(deps)* Bump anchore/sbom-action from 0.20.0 to 0.20.1 (#194) (#194)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#198) (#198)
- *(deps)* Bump anchore/sbom-action from 0.20.1 to 0.20.2 (#199) (#199)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#201) (#201)
- *(deps)* Bump anchore/sbom-action from 0.20.2 to 0.20.4 (#203) (#203)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#204) (#204)
- *(deps)* Bump google.golang.org/protobuf from 1.36.6 to 1.36.7 (#205) (#205)
- *(deps)* Bump actions/checkout from 4 to 5 (#206) (#206)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#208) (#208)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#247) (#247)
- *(deps)* Bump github/codeql-action from 3 to 4 (#241) (#241)
- *(deps)* Bump actions/stale from 9 to 10 (#224) (#224)
- *(deps)* Bump goreleaser/goreleaser-action from 6.3.0 to 6.4.0 (#210) (#210)
- *(deps)* Bump anchore/sbom-action from 0.20.4 to 0.20.6 (#232) (#232)
- *(deps)* Bump actions/setup-go from 5 to 6 (#223) (#223)
- *(deps)* Bump golang.org/x/time from 0.12.0 to 0.14.0 (#244) (#244)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#249) (#249)
- *(deps)* Bump the hashicorp group with 2 updates (#248) (#248)
- *(deps)* Bump anchore/sbom-action from 0.20.6 to 0.20.8 (#251) (#251)

### <!-- 9 -->Chore

- Fix .goreleaser deprecation warnings
- Fix ST1017: don't use Yoda conditions (staticcheck)
- Fix workflow to test the whole code

## [0.8.0] - 2025-04-04


### <!-- 0 -->Features

- Upgrade the plugin to Gitlab 17.10

### <!-- 1 -->Bug Fixes

- Wrong data returned for some token types instead of Data it was returning the Internal data

### <!-- 5 -->Tests

- Restructure code for test

### <!-- 7 -->Dependencies

- *(deps)* Bump golangci/golangci-lint-action from 6 to 7 (#175) (#175)
- *(deps)* Bump google.golang.org/protobuf from 1.36.5 to 1.36.6 (#174) (#174)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#177) (#177)
- *(deps)* Bump goreleaser/goreleaser-action from 6.2.1 to 6.3.0 (#176) (#176)

### <!-- 9 -->Chore

- Bump go to 1.24

## [0.7.4] - 2025-03-20


### <!-- 1 -->Bug Fixes

- Missing scope from group and project deploy tokens
- Separate tokens based on their types and add missing properties (#173)

## [0.7.3] - 2025-03-19


### <!-- 0 -->Features

- Allow the flags to be configurable during runtime

## [0.7.2] - 2025-03-10


### <!-- 0 -->Features

- Add flags to the plugin and implement show config token flag (#168) (#168)

### <!-- 4 -->Documentation

- Update README with upgrade note for 0.7.x
- Clarify upgrade from 0.7.x

### <!-- 7 -->Dependencies

- *(deps)* Bump google.golang.org/protobuf from 1.35.2 to 1.36.0 (#142) (#142)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#143) (#143)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#144) (#144)
- *(deps)* Bump google.golang.org/protobuf from 1.36.0 to 1.36.1 (#145) (#145)
- *(deps)* Bump golang.org/x/time from 0.8.0 to 0.9.0 (#146) (#146)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#147) (#147)
- *(deps)* Bump google.golang.org/protobuf from 1.36.1 to 1.36.2 (#148) (#148)
- *(deps)* Bump github.com/hashicorp/vault/sdk in the hashicorp group (#149) (#149)
- *(deps)* Bump google.golang.org/protobuf from 1.36.2 to 1.36.3 (#151) (#151)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#152) (#152)
- *(deps)* Bump anchore/sbom-action from 0.17.9 to 0.18.0 (#153) (#153)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#155) (#155)
- *(deps)* Bump golang.org/x/time from 0.9.0 to 0.10.0 (#156) (#156)
- *(deps)* Bump google.golang.org/protobuf from 1.36.3 to 1.36.4 (#154) (#154)
- *(deps)* Bump gitlab.com/gitlab-org/api/client-go (#162) (#162)
- *(deps)* Bump goreleaser/goreleaser-action from 6.1.0 to 6.2.1 (#160) (#160)
- *(deps)* Bump google.golang.org/protobuf from 1.36.4 to 1.36.5 (#158) (#158)
- *(deps)* Bump the hashicorp group across 1 directory with 2 updates (#164) (#164)
- *(deps)* Bump golang.org/x/time from 0.10.0 to 0.11.0 (#166) (#166)

### <!-- 9 -->Chore

- Bundle hashicorp dependabot updates
- Update go to 1.23
- Update go to 1.23.4

## [0.7.1] - 2024-12-18


### <!-- 0 -->Features

- Implement trigger and project/group deploy tokens (#140)

### <!-- 1 -->Bug Fixes

- Description for config/<name>/rotate
- Correct gitlab type references (#125)

### <!-- 4 -->Documentation

- Update README with rotate token information

### <!-- 7 -->Dependencies

- *(deps)* Bump anchore/sbom-action from 0.17.3 to 0.17.4 (#124) (#124)
- *(deps)* Bump anchore/sbom-action from 0.17.4 to 0.17.5 (#126) (#126)
- *(deps)* Bump anchore/sbom-action from 0.17.5 to 0.17.6 (#127) (#127)
- *(deps)* Bump anchore/sbom-action from 0.17.6 to 0.17.7 (#129) (#129)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.112.0 to 0.113.0 (#128) (#128)
- *(deps)* Bump goreleaser/goreleaser-action from 6.0.0 to 6.1.0 (#130) (#130)
- *(deps)* Bump github.com/stretchr/testify from 1.9.0 to 1.10.0 (#136) (#136)
- *(deps)* Bump anchore/sbom-action from 0.17.7 to 0.17.8 (#135) (#135)
- *(deps)* Bump codecov/codecov-action from 4 to 5 (#133) (#133)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.113.0 to 0.114.0 (#134) (#134)
- *(deps)* Bump golang.org/x/time from 0.7.0 to 0.8.0 (#131) (#131)
- *(deps)* Bump gopkg.in/dnaeon/go-vcr.v4 (#137) (#137)
- *(deps)* Bump google.golang.org/protobuf from 1.35.1 to 1.35.2 (#132) (#132)
- *(deps)* Bump anchore/sbom-action from 0.17.8 to 0.17.9 (#141) (#141)

### <!-- 9 -->Chore

- Added workflow for stale issues and pr

## [0.6.1] - 2024-10-15


### <!-- 1 -->Bug Fixes

- Display wrong expiry date on the rotated token
- Dont write the token config as a secret

### <!-- 4 -->Documentation

- Fix wrong property "config" to "config_name"

### <!-- 5 -->Tests

- Pass time values in ctx so we can fix the value during tests (#121)

### <!-- 7 -->Dependencies

- *(deps)* Bump github.com/xanzy/go-gitlab from 0.111.0 to 0.112.0 (#122) (#122)

### Examples

- Some examples of terraform for the vault plugin

## [0.6.0] - 2024-10-12


### <!-- 0 -->Features

- Allow us to specify multiple configurations and config per role (#120)

### <!-- 1 -->Bug Fixes

- Update endpoint names to use framework.GenericNameRegex

### <!-- 4 -->Documentation

- Removed double parenthesis in functions list

### <!-- 7 -->Dependencies

- *(deps)* Bump anchore/sbom-action from 0.17.2 to 0.17.3 (#119) (#119)

## [0.5.0] - 2024-10-12


### <!-- 0 -->Features

- Allow dynamic naming of GitLab tokens using the name property
- Add User and Group Service Accounts
- Allow you to patch every config property as needed (#118)

### <!-- 7 -->Dependencies

- *(deps)* Bump anchore/sbom-action from 0.17.0 to 0.17.2 (#109) (#109)

## [0.4.1] - 2024-07-22


### <!-- 8 -->Build

- Added illumos build to goreleaser

## [0.4.0] - 2024-07-16


### <!-- 0 -->Features

- Token rotation will use its own endpoint and tests will run against self-hosted GitLab CE (#97)

### <!-- 4 -->Documentation

- Added upgrade instructions

### <!-- 7 -->Dependencies

- *(deps)* Bump github.com/xanzy/go-gitlab from 0.102.0 to 0.103.0 (#75) (#75)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.103.0 to 0.104.0 (#80) (#80)
- *(deps)* Bump golangci/golangci-lint-action from 4 to 6 (#83) (#83)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.104.0 to 0.105.0 (#85) (#85)
- *(deps)* Bump goreleaser/goreleaser-action from 5.0.0 to 5.1.0 (#84) (#84)
- *(deps)* Bump anchore/sbom-action from 0.15.10 to 0.15.11 (#78) (#78)
- *(deps)* Bump anchore/sbom-action from 0.16.0 to 0.16.1 (#96) (#96)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.105.0 to 0.106.0 (#95) (#95)
- *(deps)* Bump goreleaser/goreleaser-action from 5.1.0 to 6.0.0 (#93) (#93)
- *(deps)* Bump anchore/sbom-action from 0.16.1 to 0.17.0 (#98) (#98)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.106.0 to 0.107.0 (#100) (#100)
- *(deps)* Bump anchore/sbom-action from 0.16.1 to 0.17.0 (#101) (#101)

### <!-- 8 -->Build

- Deprecated --rm-dist for goreleaser

### <!-- 9 -->Chore

- Remove SLSA from the workflow

## [0.3.3] - 2024-04-13


### <!-- 7 -->Dependencies

- *(deps)* Bump github.com/xanzy/go-gitlab from 0.101.0 to 0.102.0 (#70) (#70)
- *(deps)* Bump github.com/hashicorp/vault/sdk from 0.11.1 to 0.12.0 (#73) (#73)
- *(deps)* Bump slsa-framework/slsa-github-generator (#71) (#71)
- *(deps)* Bump slsa-framework/slsa-verifier from 2.4.0 to 2.5.1 (#72) (#72)

### <!-- 8 -->Build

- Fix release with SLSA

## [0.3.2] - 2024-04-03


### <!-- 1 -->Bug Fixes

- Show revoke_auto_rotated_token on read config

### <!-- 7 -->Dependencies

- *(deps)* Bump github.com/hashicorp/go-hclog from 1.6.2 to 1.6.3 (#69) (#69)

## [0.3.1] - 2024-03-29


### <!-- 1 -->Bug Fixes

- Expiry date for main token, and documentation for rotating token

### <!-- 7 -->Dependencies

- *(deps)* Bump github.com/stretchr/testify from 1.8.4 to 1.9.0 (#54) (#54)
- *(deps)* Bump anchore/sbom-action from 0.15.8 to 0.15.9 (#56) (#56)
- *(deps)* Bump slsa-framework/slsa-github-generator (#64) (#64)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.97.0 to 0.100.0 (#58) (#58)
- *(deps)* Bump github.com/hashicorp/vault/sdk from 0.11.0 to 0.11.1 (#60) (#60)
- *(deps)* Bump github.com/hashicorp/vault/api from 1.12.0 to 1.12.2 (#63) (#63)
- *(deps)* Bump google.golang.org/protobuf from 1.32.0 to 1.33.0 (#57) (#57)
- *(deps)* Bump anchore/sbom-action from 0.15.9 to 0.15.10 (#67) (#67)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.100.0 to 0.101.0 (#66) (#66)
- *(deps)* Bump slsa-framework/slsa-verifier from 2.4.1 to 2.5.1 (#65) (#65)

## [0.3.0] - 2024-02-23


### <!-- 0 -->Features

- Add gitlab revoke option so vault doesn't revoke the token when the parent expires (#52)

### <!-- 1 -->Bug Fixes

- Deprecated functions from gitlab package
- Deprecated functions from gitlab package g.Time to g.Ptr

### <!-- 5 -->Tests

- Added tests for utils convertToInt

### <!-- 7 -->Dependencies

- *(deps)* Bump golang.org/x/time from 0.3.0 to 0.4.0 (#25) (#25)
- *(deps)* Bump slsa-framework/slsa-verifier from 2.4.0 to 2.4.1 (#26) (#26)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.93.2 to 0.94.0 (#27) (#27)
- *(deps)* Bump anchore/sbom-action from 0.14.3 to 0.15.0 (#28) (#28)
- *(deps)* Bump golang.org/x/time from 0.4.0 to 0.5.0 (#29) (#29)
- *(deps)* Bump github.com/hashicorp/go-hclog from 1.5.0 to 1.6.1 (#31) (#31)
- *(deps)* Bump anchore/sbom-action from 0.15.0 to 0.15.1 (#30) (#30)
- *(deps)* Bump actions/setup-go from 4 to 5 (#32) (#32)
- *(deps)* Bump github/codeql-action from 2 to 3 (#34) (#34)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.94.0 to 0.95.2 (#35) (#35)
- *(deps)* Bump github.com/hashicorp/go-hclog from 1.6.1 to 1.6.2 (#36) (#36)
- *(deps)* Bump google.golang.org/protobuf from 1.31.0 to 1.32.0 (#37) (#37)
- *(deps)* Bump anchore/sbom-action from 0.15.1 to 0.15.2 (#38) (#38)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.95.2 to 0.96.0 (#40) (#40)
- *(deps)* Bump anchore/sbom-action from 0.15.2 to 0.15.4 (#41) (#41)
- *(deps)* Bump anchore/sbom-action from 0.15.4 to 0.15.6 (#44) (#44)
- *(deps)* Bump github.com/hashicorp/vault/api from 1.10.0 to 1.11.0 (#43) (#43)
- *(deps)* Bump anchore/sbom-action from 0.15.6 to 0.15.7 (#45) (#45)
- *(deps)* Bump anchore/sbom-action from 0.15.7 to 0.15.8 (#46) (#46)
- *(deps)* Bump codecov/codecov-action from 3 to 4 (#47) (#47)
- *(deps)* Bump golangci/golangci-lint-action from 3 to 4 (#51) (#51)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.96.0 to 0.97.0 (#48) (#48)
- *(deps)* Bump github.com/hashicorp/vault/api from 1.11.0 to 1.12.0 (#50) (#50)
- *(deps)* Bump github.com/hashicorp/vault/sdk from 0.10.2 to 0.11.0 (#49) (#49)

### <!-- 9 -->Chore

- Gofmt -w -r 'interface{} -> any' *.go
- Goimports -w *.go'

## [0.2.4] - 2023-11-06


### <!-- 9 -->Chore

- Fix goreleaser to add v to the version of the plugin as Vault expectes it with a v prefix

## [0.2.3] - 2023-10-31


### <!-- 4 -->Documentation

- Update README with example of how to user service accounts in Gitlab 16.1
- Added secrets/tune command to the example to expand MaxTTL for the mount

### <!-- 7 -->Dependencies

- *(deps)* Bump github.com/xanzy/go-gitlab from 0.91.1 to 0.92.1 (#16) (#16)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.92.1 to 0.92.3 (#17) (#17)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.92.3 to 0.93.0 (#18) (#18)
- *(deps)* Bump github.com/hashicorp/vault/sdk from 0.10.0 to 0.10.1 (#20) (#20)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.93.0 to 0.93.1 (#19) (#19)
- *(deps)* Bump github.com/hashicorp/vault/sdk from 0.10.1 to 0.10.2 (#21) (#21)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.93.1 to 0.93.2 (#22) (#22)

## [0.2.2] - 2023-09-19


### <!-- 1 -->Bug Fixes

- Auto rotate token should be between DefaultAutoRotateBeforeMinTTL and DefaultAutoRotateBeforeMaxTTL

## [0.2.1] - 2023-09-19


### <!-- 1 -->Bug Fixes

- Version should now correctly be reported after build in Vault

### <!-- 5 -->Tests

- Fix mockEventsSender to implement SendEvent instead of Send due to package upgrade

### <!-- 7 -->Dependencies

- *(deps)* Bump actions/checkout from 3 to 4 (#5) (#5)
- *(deps)* Bump goreleaser/goreleaser-action from 4.4.0 to 4.6.0 (#6) (#6)
- *(deps)* Bump goreleaser/goreleaser-action from 4.6.0 to 5.0.0 (#8) (#8)
- *(deps)* Bump codecov/codecov-action from 3 to 4 (#9) (#9)
- *(deps)* Bump github.com/hashicorp/vault/api from 1.9.2 to 1.10.0 (#10) (#10)
- *(deps)* Bump github.com/xanzy/go-gitlab from 0.90.0 to 0.91.1 (#12) (#12)
- *(deps)* Bump github.com/hashicorp/vault/sdk from 0.9.2 to 0.10.0 (#11) (#11)

### <!-- 9 -->Chore

- *(docs)* Fix incorrect base_url in gitlab/config (#13) (#13)

## [0.2.0] - 2023-09-03


### <!-- 0 -->Features

- Auto rotate the main configuration token (#4)
- Add a version to the plugin

### <!-- 5 -->Tests

- Remove getVcr commented function

## [0.1.1] - 2023-08-31


### <!-- 0 -->Features

- Added a delete operation for the config backend

### <!-- 4 -->Documentation

- Update the links for gitlab access tokens
- Update README with example of how to configure and use the plugin

### <!-- 7 -->Dependencies

- *(deps)* Bump slsa-framework/slsa-verifier from 2.1.0 to 2.4.0
- *(deps)* Bump slsa-framework/slsa-github-generator

### <!-- 9 -->Chore

- *(github)* Removed go-version-file flag in setup-go action

## [0.1.0] - 2023-08-30


### <!-- 3 -->Refactor

- Removed unnecessary version package

### <!-- 4 -->Documentation

- Update badges in README to redirect to targets

### <!-- 8 -->Build

- Added goreleaser and workflow to generate binary releases
- Fix release workflow to download syft

<!-- generated by git-cliff -->
