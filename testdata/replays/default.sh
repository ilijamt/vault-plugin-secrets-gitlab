#!/usr/bin/env bash

set -eux

vault write gitlab/config/default base_url=http://localhost:9000 token=glpat-secret-random-token auto_rotate_token=false auto_rotate_before=48h type=self-managed
vault write gitlab/roles/normal-user name='{{ .role_name }}-{{ .token_type }}-test' path=normal-user scopes="read_api" token_type=personal ttl=48h
TMPFILE=$(mktemp)
vault read -format=yaml gitlab/token/normal-user | tee "$TMPFILE"
curl --fail -H "Private-Token: $(cat $TMPFILE | yq .data.token)" http://localhost:9000/api/v4/personal_access_tokens/self
vault lease revoke -prefix gitlab/token/normal-user
vault write -f gitlab/config/default/rotate
vault read -format=yaml gitlab/token/normal-user | tee "$TMPFILE"
curl --fail -H "Private-Token: $(cat $TMPFILE | yq .data.token)" http://localhost:9000/api/v4/personal_access_tokens/self
vault lease revoke -prefix gitlab/token/normal-user
curl -H "Private-Token: $(cat $TMPFILE | yq .data.token)" http://localhost:9000/api/v4/personal_access_tokens/self
rm "$TMPFILE"