# Terraform is a general purpose tool. To interact with specific APIs, it
# requires users to configure plugins called "providers".
# For more information: https://www.terraform.io/docs/language/providers/index.html

# The "azurerm" provider enables us to provision cloud resources on Azure.
provider "azurerm" {
  features {}

  # "Default subscription" subscription
  subscription_id = "8ab3da27-5e1b-494f-abc6-726fb04729b3"

  # We don't need to register any resource providers.
  # For more information: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#skip_provider_registration
  skip_provider_registration = false
}

# The "azuread" provider enables us to provision resources in Azure Active
# Directory.
provider "azuread" {
  # Default Directory
  tenant_id = "0581e2b2-19ee-4e7c-94f7-d3e38a2409df"
}

# The "github" provider enables us to configure CI/CD on GitHub.
provider "github" {}
