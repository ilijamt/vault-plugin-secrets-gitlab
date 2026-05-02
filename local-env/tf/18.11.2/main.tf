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

module "shared" {
  source = "../_shared"
}
