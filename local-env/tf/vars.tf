locals {
  token_expiry_time = formatdate("YYYY-MM-DD", time_offset.one_year_later.rfc3339)
  scopes_admin_user = ["sudo", "admin_mode"]
  scopes_user_token = ["api"]
}



