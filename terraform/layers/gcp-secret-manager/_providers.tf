# Terraform is a general purpose tool. To interact with specific APIs, it
# requires users to configure plugins called "providers".
# For more information: https://www.terraform.io/docs/language/providers/index.html

# The "google" provider enables us to provision cloud resources on Google Cloud
# Platform.
provider "google" {
  project = "whisper-tests"
  region  = "europe-west9"
}

# The "google-beta" provider enables us to use features of Google Cloud Platform
# that are still in beta. The use of beta features should generally be kept to a
# minimum, but Google's betas are overall very stable.
provider "google-beta" {
  project = "whisper-tests"
  region  = "europe-west9"
}

# The "github" provider enables us to configure CI/CD on GitHub.
provider "github" {}
