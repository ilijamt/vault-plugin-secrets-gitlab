terraform {
  required_providers {
    gitlab = {
      source = "gitlabhq/gitlab"
      version = "17.1.0"
    }
  }
}

provider "gitlab" {
  base_url = "http://localhost:8080"
  insecure = true
  token = "glpat-secret-random-token"
}
