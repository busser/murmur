resource "google_service_account" "github_actions" {
  account_id = "github-actions"

  display_name = "Github Actions"
}

resource "google_secret_manager_secret_iam_member" "github_actions_access_secret" {
  secret_id = google_secret_manager_secret.example.secret_id

  role   = "roles/secretmanager.secretAccessor"
  member = "serviceAccount:${google_service_account.github_actions.email}"
}

# We use workload identity to enable keyless authentication from whisper's
# Github Actions workflows.
resource "google_iam_workload_identity_pool" "default" {
  provider = google-beta

  workload_identity_pool_id = "default"
}

resource "google_iam_workload_identity_pool_provider" "github_oidc" {
  provider = google-beta
  project  = local.google_project

  workload_identity_pool_provider_id = "github-oidc"

  workload_identity_pool_id = google_iam_workload_identity_pool.default.workload_identity_pool_id
  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.actor"      = "assertion.actor"
    "attribute.aud"        = "assertion.aud"
    "attribute.repository" = "assertion.repository"
  }

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

# Whisper's Github Actions workflows use a dedicated Google service account
# to interact with the Google API and access secret versions.
resource "google_service_account_iam_member" "github_actions_workload_identity" {
  service_account_id = google_service_account.github_actions.id

  role   = "roles/iam.workloadIdentityUser"
  member = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.default.name}/attribute.repository/busser/whisper"
}
