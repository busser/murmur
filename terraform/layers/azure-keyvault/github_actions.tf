// The repository's continuous integration pipelines run whisper's end-to-end
// tests. These tests require credentials that can read secrets from our Key
// Vaults.

// The pipelines authenticate to Azure with a service principal.

resource "azuread_application" "github_actions" {
  display_name = "whisper-github-actions"
  owners       = [data.azuread_client_config.current.object_id]
}

resource "azuread_service_principal" "github_actions" {
  application_id               = azuread_application.github_actions.application_id
  app_role_assignment_required = false
  owners                       = [data.azuread_client_config.current.object_id]
}

resource "azuread_service_principal_password" "github_actions" {
  service_principal_id = azuread_service_principal.github_actions.object_id
}

// The necessary credentials are stored in this repository's Github Actions
// secrets. Pipelines use these secrets to set environment variables used by
// whisper.

data "github_repository" "whisper" {
  name = "whisper"
}

resource "github_actions_secret" "tenant_id" {
  repository      = data.github_repository.whisper.name
  secret_name     = "AZURE_TENANT_ID"
  plaintext_value = data.azuread_client_config.current.tenant_id
}

resource "github_actions_secret" "client_id" {
  repository      = data.github_repository.whisper.name
  secret_name     = "AZURE_CLIENT_ID"
  plaintext_value = azuread_service_principal.github_actions.application_id
}

resource "github_actions_secret" "client_secret" {
  repository      = data.github_repository.whisper.name
  secret_name     = "AZURE_CLIENT_SECRET"
  plaintext_value = azuread_service_principal_password.github_actions.value
}
