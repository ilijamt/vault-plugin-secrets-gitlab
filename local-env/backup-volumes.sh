#!/usr/bin/env bash

set -x

docker compose stop
docker run --rm --volumes-from vpsg-web-1 -v $(pwd):/backup ubuntu tar cvf /backup/backup.tar /etc/gitlab /var/opt/gitlab/postgresql/
docker compose up -d