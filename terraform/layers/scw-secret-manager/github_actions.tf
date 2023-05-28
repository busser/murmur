data "scaleway_account_project" "current" {
  project_id = "5e83ac90-5df2-4c7d-98ba-ef50ff4d148a" # Cloudlab
}

resource "scaleway_iam_application" "github_actions" {
  name        = "murmur-github-actions"
  description = "Github Actions (busser/murmur)"
}

resource "scaleway_iam_api_key" "github_actions" {
  application_id = scaleway_iam_application.github_actions.id
  description    = "Used by Github Actions (busser/murmur)"
}

resource "scaleway_iam_group" "secrets_readers" {
  name        = "secrets-readers"
  description = "members can read all secrets"
  application_ids = [
    scaleway_iam_application.github_actions.id
  ]
}

resource "scaleway_iam_policy" "secrets_readers" {
  name        = "secrets-readers"
  description = "grants read-only access to all secrets"
  group_id    = scaleway_iam_group.secrets_readers.id
  rule {
    project_ids = [
      data.scaleway_account_project.current.id,
    ]
    permission_set_names = [
      # Must grant full access because the ReadOnly permission set does not
      # grant access to secret values.
      # Discussion started on Slack here:
      # https://scaleway-community.slack.com/archives/C04KGMME3U1/p1682867139987979
      "SecretManagerFullAccess",
      # "SecretManagerReadOnly",
    ]
  }
}

// The necessary credentials are stored in this repository's Github Actions
// secrets. Pipelines use these secrets to set environment variables used by
// murmur.

data "github_repository" "murmur" {
  name = "murmur"
}

resource "github_actions_secret" "access_key" {
  repository      = data.github_repository.murmur.name
  secret_name     = "SCW_ACCESS_KEY"
  plaintext_value = scaleway_iam_api_key.github_actions.access_key
}

resource "github_actions_secret" "secret_key" {
  repository      = data.github_repository.murmur.name
  secret_name     = "SCW_SECRET_KEY"
  plaintext_value = scaleway_iam_api_key.github_actions.secret_key
}
