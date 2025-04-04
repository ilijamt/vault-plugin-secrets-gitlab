resource "local_file" "tokens" {
  filename = "${path.module}/tokens.json"
  content = jsonencode({
    admin_user_root = {
      id         = gitlab_personal_access_token.admin_user_root.id
      token      = gitlab_personal_access_token.admin_user_root.token
      created_at = gitlab_personal_access_token.admin_user_root.created_at
    }
    admin_user_initial_token = {
      id         = gitlab_personal_access_token.admin_user_initial_token.id
      token      = gitlab_personal_access_token.admin_user_initial_token.token
      created_at = gitlab_personal_access_token.admin_user_initial_token.created_at
    }
    admin_user_auto_rotate_token_main_token = {
      id         = gitlab_personal_access_token.admin_user_auto_rotate_token_main_token.id
      token      = gitlab_personal_access_token.admin_user_auto_rotate_token_main_token.token
      created_at = gitlab_personal_access_token.admin_user_auto_rotate_token_main_token.created_at
    }
    admin_user_auto_rotate_token_1 = {
      id         = gitlab_personal_access_token.admin_user_auto_rotate_token_1.id
      token      = gitlab_personal_access_token.admin_user_auto_rotate_token_1.token
      created_at = gitlab_personal_access_token.admin_user_auto_rotate_token_1.created_at
    }
    normal_user_initial_token = {
      id         = gitlab_personal_access_token.normal_user_initial_token.id
      token      = gitlab_personal_access_token.normal_user_initial_token.token
      created_at = gitlab_personal_access_token.normal_user_initial_token.created_at
    }
  })
}
