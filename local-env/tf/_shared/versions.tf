terraform {
  required_providers {
    gitlab = {
      source = "gitlabhq/gitlab"
    }
    local = {
      source = "hashicorp/local"
    }
    time = {
      source = "hashicorp/time"
    }
  }
}
