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

stage "Stopping containers"
run docker compose kill

stage "Restoring backup archive"
run docker run --rm --volumes-from vpsg-web-1 -v "$(pwd)":/backup ubuntu bash -c "rm -rf /etc/gitlab /var/opt/gitlab/postgresql/; cd / && tar xvf /backup/backup.tar"

stage "Starting containers"
run docker compose up -d

stage "Waiting for GitLab to be ready"
until curl -sSf http://localhost:8080/users/sign_in > /dev/null 2>&1; do
  echo -e "  ${CYAN}GitLab not ready yet, retrying in 10s...${RESET}"
  sleep 10
done

stage "Done"