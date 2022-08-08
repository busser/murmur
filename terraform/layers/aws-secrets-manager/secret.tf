resource "aws_secretsmanager_secret" "example" {
  name = "secret-sauce"
}

resource "aws_secretsmanager_secret_version" "ketchup" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = "ketchup"

  version_stages = [
    "v1",
  ]
}

resource "aws_secretsmanager_secret_version" "szechuan" {
  secret_id     = aws_secretsmanager_secret.example.id
  secret_string = "szechuan"

  version_stages = [
    "v2",
    "AWSCURRENT",
  ]
}
