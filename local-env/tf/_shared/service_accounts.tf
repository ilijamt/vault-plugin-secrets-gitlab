# Container group and project for the self-hosted service account tests.
resource "gitlab_group" "service_accounts" {
  name        = "service-accounts"
  path        = "service-accounts"
  description = "Container group for the service account integration tests"
}

resource "gitlab_project" "service_accounts" {
  name         = "project"
  path         = "project"
  description  = "Project hosting project-level service accounts for integration tests"
  namespace_id = gitlab_group.service_accounts.id
}
