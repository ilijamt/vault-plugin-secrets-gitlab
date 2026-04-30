resource "gitlab_personal_access_token" "admin_user_root" {
  user_id    = gitlab_user.admin_user.id
  name       = "Root token"
  expires_at = local.token_expiry_time

  scopes = concat(local.scopes_admin_user, local.scopes_user_token)
}

resource "gitlab_personal_access_token" "admin_user_initial_token" {
  user_id    = gitlab_user.admin_user.id
  name       = "Initial token"
  expires_at = local.token_expiry_time

  scopes = local.scopes_user_token
}

resource "gitlab_personal_access_token" "admin_user_auto_rotate_token_main_token" {
  user_id    = gitlab_user.admin_user.id
  name       = "Auto rotate token main token"
  expires_at = local.token_expiry_time

  scopes = local.scopes_user_token
}


resource "gitlab_personal_access_token" "admin_user_auto_rotate_token_1" {
  user_id    = gitlab_user.admin_user.id
  name       = "Auto rotate token 1"
  expires_at = local.token_expiry_time

  scopes = local.scopes_user_token
}
