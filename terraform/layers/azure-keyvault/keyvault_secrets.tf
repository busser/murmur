resource "azurerm_resource_group" "whisper" {
  name     = "whisper"
  location = "West Europe"
}

// We have multiple Key Vaults because whisper supports fetching secrets from
// multiple Key Vaults at once.
resource "azurerm_key_vault" "whisper" {
  for_each = toset(["alpha", "bravo"])

  name = "whisper-${each.key}"

  tenant_id           = data.azurerm_client_config.current.tenant_id
  location            = azurerm_resource_group.whisper.location
  resource_group_name = azurerm_resource_group.whisper.name

  soft_delete_retention_days = 7
  enable_rbac_authorization  = true

  sku_name = "standard"
}

// This secret has multiple versions because whisper supports fetching any
// version of a secret. The secret's version IDs are hard-coded in whisper's
// end-to-end tests.
resource "azurerm_key_vault_secret" "example" {
  for_each = azurerm_key_vault.whisper

  name         = "secret-sauce"
  value        = "szechuan" // Was previously applied with value "ketchup".
  key_vault_id = azurerm_key_vault.whisper[each.key].id

  depends_on = [
    azurerm_role_assignment.keyvault_admin,
  ]
}

// Infrastructure is managed by @busser. To date, he is the only person with
// write access to cloud resources used by whisper.
resource "azurerm_role_assignment" "keyvault_admin" {
  for_each = azurerm_key_vault.whisper

  scope                = azurerm_key_vault.whisper[each.key].id
  principal_id         = data.azurerm_client_config.current.object_id
  role_definition_name = "Key Vault Administrator"
}

// The repository's continuous integrations pipelines read secrets from our Key
// Vaults when running whisper's end-to-end tests.
resource "azurerm_role_assignment" "github_actions" {
  for_each = azurerm_key_vault.whisper

  scope                = azurerm_key_vault.whisper[each.key].id
  principal_id         = azuread_service_principal.github_actions.object_id
  role_definition_name = "Key Vault Secrets User"
}
