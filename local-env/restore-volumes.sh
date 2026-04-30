#!/usr/bin/env bash

set -eu

if [ "$#" -ne 1 ] || [ -z "${1:-}" ]; then
  echo "ERROR: GitLab version is required."
  echo "Usage: $0 <version>   (e.g. $0 17.11.7)"
  exit 1
fi

VERSION="$1"

if [[ ! "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "ERROR: '${VERSION}' does not look like a GitLab version (expected MAJOR.MINOR.PATCH, e.g. 17.11.7)."
  exit 1
fi

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

BACKUP_FILE="backup.${VERSION}.tar"

if [ ! -f "${BACKUP_FILE}" ]; then
  echo "ERROR: backup file '${BACKUP_FILE}' not found in $(pwd)."
  echo "Run './backup-volumes.sh ${VERSION}' first, or check the version argument."
  exit 1
fi

if ! docker inspect vpsg-web-1 >/dev/null 2>&1; then
  echo "ERROR: container 'vpsg-web-1' not found. Bring it up first:"
  echo "  GITLAB_IMAGE_TAG=${GITLAB_IMAGE_TAG} docker compose up -d --wait"
  exit 1
fi

stage "Stopping containers"
run docker compose kill

stage "Restoring backup archive ${BACKUP_FILE}"
run docker run --rm --volumes-from vpsg-web-1 -v "$(pwd)":/backup ubuntu \
  bash -c "rm -rf /etc/gitlab /var/opt/gitlab/postgresql/; cd / && tar xvf /backup/${BACKUP_FILE}"

stage "Starting containers"
run docker compose up -d

stage "Waiting for GitLab to be ready"
until curl -sSf http://localhost:8080/users/sign_in > /dev/null 2>&1; do
  echo -e "  ${CYAN}GitLab not ready yet, retrying in 10s...${RESET}"
  sleep 10
done

stage "Done"
