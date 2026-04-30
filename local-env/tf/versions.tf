terraform {
  required_providers {
    gitlab = {
      source  = "gitlabhq/gitlab"
      version = "~> 17.10"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.13"
    }
  }
}

variable "gitlab_base_url" {
  type    = string
  default = "http://localhost:8080"
}

variable "gitlab_token" {
  type    = string
  default = "glpat-secret-random-token"
}

provider "gitlab" {
  base_url = var.gitlab_base_url
  insecure = true
  token    = var.gitlab_token
}
