resource "google_secret_manager_secret" "example" {
  secret_id = "secret-sauce"

  replication {
    automatic = true
  }

  depends_on = [
    google_project_service.secretmanager,
  ]
}

resource "google_secret_manager_secret_version" "ketchup" {
  secret = google_secret_manager_secret.example.id

  secret_data = "ketchup"
}

resource "google_secret_manager_secret_version" "szechuan" {
  secret = google_secret_manager_secret.example.id

  secret_data = "szechuan"

  depends_on = [
    // This controls the order in which the versions are created.
    google_secret_manager_secret_version.ketchup,
  ]
}
