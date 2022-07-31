# Terraform is a general purpose tool. To interact with specific APIs, it
# requires users to configure plugins called "providers".
# For more information: https://www.terraform.io/docs/language/providers/index.html

# The "scaleway" provider enables us to provision cloud resources on Scaleway.
provider "scaleway" {
  project_id = "5e83ac90-5df2-4c7d-98ba-ef50ff4d148a" # Cloudlab
  region     = "fr-par"
  zone       = "fr-par-1"
}
