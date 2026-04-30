#!/usr/bin/env bash

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

BACKUP_FILE="backup.${VERSION}.tar"

stage "Removing old backup ${BACKUP_FILE}"
run rm -f "${BACKUP_FILE}"

stage "Stopping containers"
run docker compose stop

stage "Creating backup archive ${BACKUP_FILE}"
run docker run --rm --volumes-from vpsg-web-1 -v "$(pwd)":/backup ubuntu \
  tar cvf "/backup/${BACKUP_FILE}" /etc/gitlab /var/opt/gitlab/postgresql/ > /dev/null

stage "Starting containers"
run docker compose up -d

stage "Done"
