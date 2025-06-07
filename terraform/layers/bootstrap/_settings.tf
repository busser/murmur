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

  # This layer's state is stored locally and persisted in the git repository.
  backend "local" {}

  # This layer requires that certain providers be configured by the caller.
  # For more information: https://www.terraform.io/docs/language/providers/requirements.html
  required_providers {
    scaleway = {
      source  = "scaleway/scaleway"
      version = "2.55.0"
    }
  }
}
