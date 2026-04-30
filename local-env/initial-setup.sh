#!/bin/bash

set -eu

VERSION="${1:-17.11.7}"
export GITLAB_IMAGE_TAG="${VERSION}-ce.0"

BOLD='\033[1m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
RESET='\033[0m'

stage() {
  echo ""
  echo -e "${BOLD}${GREEN}=== [$1] ===${RESET}"
  echo ""
}

run() {
  echo -e "  ${CYAN}\$ $*${RESET}" >&2
  "$@"
}

TOKENS_OUT="../tests/integration/testdata/tokens.${VERSION}.json"
TF_DIR="tf/${VERSION}"

if [ ! -d "${TF_DIR}" ]; then
  echo "ERROR: no terraform directory at ${TF_DIR} for version ${VERSION}. Create it (copy from an existing tf/<version>/) and adjust the gitlab provider version in versions.tf."
  exit 1
fi

fn() {
  local cmd=$1
  echo "$cmd"
  docker exec vpsg-web-1 gitlab-rails runner "$cmd"
}

is_gitlab_up() {
  curl -sSf http://localhost:8080/users/sign_in > /dev/null 2>&1
}

stage "Tearing down existing environment"
run docker compose kill
run docker compose down --remove-orphans --volumes

stage "Starting containers for GitLab ${VERSION} (image gitlab/gitlab-ce:${GITLAB_IMAGE_TAG})"
run docker compose up -d --wait
run rm -rf "${TF_DIR}/.terraform" "${TF_DIR}/terraform.tfstate"*
run docker compose up -d --wait

stage "Waiting for GitLab to be ready"
until is_gitlab_up; do
  echo -e "  ${CYAN}GitLab not ready yet, retrying in 10s...${RESET}"
  sleep 10
done

stage "Creating initial admin access token"
fn 'token = User.find_by_username("root").personal_access_tokens.create(name: "Initial token", expires_at: DateTime.now.next_month(6).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-random-token"); token.save!'

stage "Running Terraform from ${TF_DIR}"
cd "${TF_DIR}" || exit
run terraform init
run terraform apply --auto-approve
cd - >/dev/null

stage "Saving access tokens to testdata (${TOKENS_OUT})"
cat "${TF_DIR}/tokens.json" | jq . > "${TOKENS_OUT}"

stage "Backing up volumes"
run bash ./backup-volumes.sh "${VERSION}"

stage "Waiting for GitLab to be ready"
until is_gitlab_up; do
  echo -e "  ${CYAN}GitLab not ready yet, retrying in 10s...${RESET}"
  sleep 10
done

stage "Done"
