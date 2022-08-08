# Terraform is a general purpose tool. To interact with specific APIs, it
# requires users to configure plugins called "providers".
# For more information: https://www.terraform.io/docs/language/providers/index.html

# The "aws" provider enables us to provision cloud resources on Amazon Web
# Services.
provider "aws" {}

# The "github" provider enables us to configure CI/CD on GitHub.
provider "github" {}
