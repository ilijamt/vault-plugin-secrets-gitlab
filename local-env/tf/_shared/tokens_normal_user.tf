resource "gitlab_personal_access_token" "normal_user_initial_token" {
  user_id    = gitlab_user.normal_user.id
  name       = "Initial token"
  expires_at = local.token_expiry_time

  scopes = local.scopes_user_token
}
