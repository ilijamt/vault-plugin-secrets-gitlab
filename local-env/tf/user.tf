resource "gitlab_user" "admin_user" {
  name           = "Admin User"
  username       = "admin-user"
  password       = "quaijooMeewieMieM1bi"
  email          = "admin@local"
  is_admin       = true
  reset_password = false
}

resource "gitlab_user" "normal_user" {
  name           = "Normal User"
  username       = "normal-user"
  password       = "cashaep4ONgahCae0bae"
  email          = "normal@local"
  is_admin       = false
  reset_password = false
}
