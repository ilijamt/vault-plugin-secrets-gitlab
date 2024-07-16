#!/usr/bin/env bash

docker compose kill
docker run --rm --volumes-from vpsg-web-1 -v $(pwd):/backup ubuntu bash -c "rm -rf /etc/gitlab /var/opt/gitlab/postgresql/; cd / && tar xvf /backup/backup.tar"
docker compose up -d