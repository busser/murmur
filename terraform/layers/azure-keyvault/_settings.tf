terraform {
  # At the root of a layer (ie, the directory where "terraform apply" is run),
  # best practice is to specify an exact version of Terraform to use. Use the
  # "= 1.2.3" constraint to do this.
  #
  # In a module, you can allow more flexibility with regards to Terraform's
  # minor and/or patch versions. For example, the "~> 1.0" constraint will allow
  # all 1.x.x versions of Terraform, while the "~> 1.0.0" constraint will allow
  # all 1.0.x versions.
  #
  # For more information: https://www.terraform.io/docs/language/settings/index.html#specifying-a-required-terraform-version
  required_version = "~> 1.4"

  # Terraform keeps track of all resources it knows of in its state. This state
  # can be stored remotely in a "backend".
  # For more information on state backends: https://www.terraform.io/docs/language/settings/backends/index.html
  # For more information on the "s3" backend: https://www.terraform.io/docs/language/settings/backends/s3.html
  backend "s3" {
    bucket   = "b4r-whisper-tfstate"
    key      = "azurerm-keyvault"
    region   = "fr-par"
    endpoint = "https://s3.fr-par.scw.cloud"
    profile  = "scaleway"
    # We are swapping the AWS S3 API for the Scaleway S3 API, so we need to 
    # skip certain validation steps.
    skip_credentials_validation = true
    skip_region_validation      = true
  }

  # This layer requires that certain providers be configured by the caller.
  # For more information: https://www.terraform.io/docs/language/providers/requirements.html
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.50.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "2.31.0"
    }
    github = {
      source  = "integrations/github"
      version = "5.25.1"
    }
  }
}
