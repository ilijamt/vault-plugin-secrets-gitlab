#!/usr/bin/env bash
set -euo pipefail

: "${IMAGE:?IMAGE required (e.g. ghcr.io/owner/name)}"
: "${TAG:?TAG required (e.g. v1.2.3)}"
: "${AMD64_BIN:?AMD64_BIN required (path to linux/amd64 binary)}"
: "${ARM64_BIN:?ARM64_BIN required (path to linux/arm64 binary)}"

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
dockerfile="${repo_root}/Dockerfile"

build_arch() {
  local platform="$1" binpath="$2" image_tag="$3" ctx
  [ -f "$binpath" ] || { echo "binary not found: $binpath" >&2; exit 1; }
  ctx="$(mktemp -d)"
  trap 'rm -rf "$ctx"' RETURN
  cp "$binpath" "${ctx}/vault-plugin-secrets-gitlab"
  docker buildx build \
    --platform "$platform" \
    --file "$dockerfile" \
    --tag "$image_tag" \
    --push \
    "$ctx"
}

build_arch linux/amd64 "$AMD64_BIN" "${IMAGE}:${TAG}-amd64"
build_arch linux/arm64 "$ARM64_BIN" "${IMAGE}:${TAG}-arm64"

docker buildx imagetools create --tag "${IMAGE}:${TAG}" \
  "${IMAGE}:${TAG}-amd64" "${IMAGE}:${TAG}-arm64"

if [ "${ALSO_LATEST:-false}" = "true" ]; then
  docker buildx imagetools create --tag "${IMAGE}:latest" \
    "${IMAGE}:${TAG}-amd64" "${IMAGE}:${TAG}-arm64"
fi

echo "Pushed ${IMAGE}:${TAG} (linux/amd64, linux/arm64)"
