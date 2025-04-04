#/bin/bash

set -x
docker compose kill
docker compose down --remove-orphans --volumes
docker compose up -d --wait
rm -f tf/terraform.tfstate*
docker compose up -d --wait

fn() {
  local cmd=$1
  echo "$cmd"
  docker exec vpsg-web-1 gitlab-rails runner "$cmd"
}

is_gitlab_up() {
  curl -sSf http://localhost:8080/users/sign_in > /dev/null 2>&1
}

# Wait for GitLab to be up
until is_gitlab_up; do
  echo "Waiting for GitLab to be up..."
  sleep 10
done

fn 'token = User.find_by_username("root").personal_access_tokens.create(name: "Initial token", expires_at: DateTime.now.next_month(6).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-random-token"); token.save!'

set -eux

cd tf || exit
terraform init
terraform apply --auto-approve
cat tokens.json  | jq . > ../../testdata/tokens.json

cd ..
bash backup-volumes.sh