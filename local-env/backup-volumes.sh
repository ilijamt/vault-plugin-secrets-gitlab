#!/usr/bin/env bash

set -eu

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

stage "Removing old backup"
run rm -f backup.tar

stage "Stopping containers"
run docker compose stop

stage "Creating backup archive"
run docker run --rm --volumes-from vpsg-web-1 -v "$(pwd)":/backup ubuntu tar cvf /backup/backup.tar /etc/gitlab /var/opt/gitlab/postgresql/ > /dev/null

stage "Starting containers"
run docker compose up -d

stage "Done"