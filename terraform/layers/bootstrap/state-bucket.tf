# Other Terraform layers' state is stored in this bucket. Each layer should use
# a different sub-path.
resource "scaleway_object_bucket" "terraform_state" {
  name = "busser-murmur-tfstate"
  versioning {
    enabled = true
  }
}
