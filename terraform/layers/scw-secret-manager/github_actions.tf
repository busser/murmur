data "scaleway_account_project" "current" {
  project_id = "5e83ac90-5df2-4c7d-98ba-ef50ff4d148a" # Cloudlab
}

resource "scaleway_iam_application" "github_actions" {
  name        = "whisper-github-actions"
  description = "Github Actions (busser/whisper)"
}

resource "scaleway_iam_api_key" "github_actions" {
  application_id = scaleway_iam_application.github_actions.id
  description    = "Used by Github Actions (busser/whisper)"
}

resource "scaleway_iam_group" "secrets_readers" {
  name        = "secrets-readers"
  description = "members can read all secrets"
  application_ids = [
    scaleway_iam_application.github_actions.id
  ]
}

resource "scaleway_iam_policy" "secrets_readers" {
  name           = "secrets-readers"
  description    = "grants read-only access to all secrets"
  application_id = scaleway_iam_application.github_actions.id
  rule {
    project_ids = [
      data.scaleway_account_project.current.id,
    ]
    permission_set_names = [
      "SecretManagerReadOnly",
    ]
  }
}

// The necessary credentials are stored in this repository's Github Actions
// secrets. Pipelines use these secrets to set environment variables used by
// whisper.

data "github_repository" "whisper" {
  name = "whisper"
}

resource "github_actions_secret" "tenant_id" {
  repository      = data.github_repository.whisper.name
  secret_name     = "SCW_ACCESS_KEY"
  plaintext_value = scaleway_iam_api_key.github_actions.access_key
}

resource "github_actions_secret" "client_id" {
  repository      = data.github_repository.whisper.name
  secret_name     = "SCW_SECRET_KEY"
  plaintext_value = scaleway_iam_api_key.github_actions.secret_key
}
