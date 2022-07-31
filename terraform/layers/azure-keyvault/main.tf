resource "azurerm_resource_group" "whisper" {
  name     = "whisper"
  location = "West Europe"
}

data "azurerm_client_config" "current" {}

resource "azurerm_key_vault" "whisper" {
  name                = "whisper"
  tenant_id           = data.azurerm_client_config.current.tenant_id
  location            = azurerm_resource_group.whisper.location
  resource_group_name = azurerm_resource_group.whisper.name

  soft_delete_retention_days = 7
  enable_rbac_authorization  = true

  sku_name = "standard"
}

resource "azurerm_role_assignment" "keyvault_admin" {
  scope                = azurerm_key_vault.whisper.id
  principal_id         = data.azurerm_client_config.current.object_id
  role_definition_name = "Key Vault Administrator"
}

resource "azurerm_key_vault_secret" "example" {
  name         = "secret-sauce"
  value        = "szechuan"
  key_vault_id = azurerm_key_vault.whisper.id

  depends_on = [
    azurerm_role_assignment.keyvault_admin,
  ]
}
