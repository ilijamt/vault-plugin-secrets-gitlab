#/bin/bash

docker compose kill
docker compose down --remove-orphans --volumes
docker compose up -d
rm -f tf/terraform.tfstate*

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
  sleep 5
done

fn 'u = User.new(admin: true, username: "admin-user", email: "admin@local", name: "Admin User", password: "quaijooMeewieMieM1bi", password_confirmation: "quaijooMeewieMieM1bi"); u.assign_personal_namespace(Organizations::Organization.default_organization); u.skip_confirmation!; u.save!'
fn 'u = User.new(username: "normal-user", email: "normal@local", name: "Normal User", password: "cashaep4ONgahCae0bae", password_confirmation: "cashaep4ONgahCae0bae"); u.assign_personal_namespace(Organizations::Organization.default_organization); u.skip_confirmation!; u.save!'
fn 'token = User.find_by_username("root").personal_access_tokens.create(name: "Initial token", expires_at: DateTime.now.next_year(1).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-random-token"); token.save!'
fn 'token = User.find_by_username("admin-user").personal_access_tokens.create(name: "Initial token", expires_at: DateTime.now.next_year(1).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-admin-token"); token.save!'
fn 'token = User.find_by_username("normal-user").personal_access_tokens.create(name: "Initial token", expires_at: DateTime.now.next_year(1).to_time, scopes: [:api, :read_api, :read_user, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-normal-token"); token.save!'
fn 'token = User.find_by_username("admin-user").personal_access_tokens.create(name: "Auto rotate token day 1", expires_at: DateTime.now.next_day(12).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-admin-token-ard1"); token.save!'
fn 'token = User.find_by_username("admin-user").personal_access_tokens.create(name: "Auto rotate token day 2", expires_at: DateTime.now.next_day(45).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-admin-token-ard2"); token.save!'
fn 'token = User.find_by_username("admin-user").personal_access_tokens.create(name: "Auto rotate token 1", expires_at: DateTime.now.next_year(1).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-admin-token-ar1"); token.save!'
fn 'token = User.find_by_username("admin-user").personal_access_tokens.create(name: "Auto rotate token 2", expires_at: DateTime.now.next_year(1).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-admin-token-ar2"); token.save!'
fn 'token = User.find_by_username("admin-user").personal_access_tokens.create(name: "Auto rotate token 3", expires_at: DateTime.now.next_year(1).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-admin-token-ar3"); token.save!'
fn 'token = User.find_by_username("normal-user").personal_access_tokens.create(name: "Auto rotate token 1", expires_at: DateTime.now.next_year(1).to_time, scopes: [:api, :read_api, :read_user, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-normal-token-ar1"); token.save!'
fn 'token = User.find_by_username("normal-user").personal_access_tokens.create(name: "Auto rotate token 2", expires_at: DateTime.now.next_year(1).to_time, scopes: [:api, :read_api, :read_user, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-normal-token-ar2"); token.save!'
fn 'token = User.find_by_username("normal-user").personal_access_tokens.create(name: "Auto rotate token day 1", expires_at: DateTime.now.next_day(45).to_time, scopes: [:api, :read_api, :read_user, :sudo, :admin_mode, :create_runner, :k8s_proxy, :read_repository, :write_repository, :ai_features, :read_service_ping]); token.set_token("glpat-secret-normal-token-ard1"); token.save!'

cd tf || exit
terraform init
terraform apply --auto-approve