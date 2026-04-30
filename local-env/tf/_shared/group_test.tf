resource "gitlab_group" "test" {
  name        = "test"
  path        = "test"
  description = "A test group"
}

resource "gitlab_group" "l1" {
  name        = "l1"
  path        = "level-1"
  description = "A root group"
}

resource "gitlab_group" "l2" {
  name        = "l2"
  path        = "level-2"
  description = "One level down"
  parent_id   = gitlab_group.l1.id
}

resource "gitlab_group" "l3" {
  name        = "l3"
  path        = "level-3"
  description = "Two levels down"
  parent_id   = gitlab_group.l2.id
}
