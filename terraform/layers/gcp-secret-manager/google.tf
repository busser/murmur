data "google_client_config" "current" {}

locals {
  google_project = data.google_client_config.current.project
}
