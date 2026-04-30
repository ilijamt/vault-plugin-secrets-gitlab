provider "time" {}

resource "time_offset" "one_year_later" {
  offset_days = 364
}
