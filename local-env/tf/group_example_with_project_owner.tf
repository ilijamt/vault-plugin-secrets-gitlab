resource "gitlab_group" "group_example" {
  name        = "example"
  path        = "example"
  description = "An example group"
}

resource "gitlab_project" "project_example" {
  name         = "example"
  description  = "An example project"
  namespace_id = gitlab_group.group_example.id
}

resource "gitlab_group_membership" "normal_user" {
  group_id     = gitlab_group.group_example.id
  user_id      = gitlab_user.normal_user.id
  access_level = "owner"
}

