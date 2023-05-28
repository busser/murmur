resource "scaleway_secret" "example" {
  name = "secret-sauce"
}

resource "scaleway_secret_version" "ketchup" {
  secret_id = scaleway_secret.example.id

  data = "ketchup"
}

resource "scaleway_secret_version" "szechuan" {
  secret_id = scaleway_secret.example.id

  data = "szechuan"

  depends_on = [
    // This controls the order in which the versions are created.
    scaleway_secret_version.ketchup,
  ]
}
