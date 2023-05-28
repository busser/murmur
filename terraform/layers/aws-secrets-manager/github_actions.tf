resource "aws_iam_openid_connect_provider" "github_actions" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list = [
    "sts.amazonaws.com",
  ]

  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1"
  ]
}

resource "aws_iam_role" "github_actions" {
  name = "GithubActions"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = [
        "sts:AssumeRoleWithWebIdentity"
      ]
      Principal = {
        Federated = aws_iam_openid_connect_provider.github_actions.arn
      }
      Condition = {
        StringLike = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com",
          "token.actions.githubusercontent.com:sub" = "repo:busser/murmur:*"
        }
      }
    }
  })
}

resource "aws_iam_role_policy" "github_actions_read_secret" {
  name = "github-actions-read-secret"
  role = aws_iam_role.github_actions.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "secretsmanager:GetSecretValue",
        ]
        Effect = "Allow"
        Resource = [
          aws_secretsmanager_secret.example.arn,
        ]
      },
    ]
  })
}

# resource "aws_secretsmanager_secret_policy" "github_actions_read_secret" {
#   secret_arn = aws_secretsmanager_secret.example.arn

#   policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = {
#       Effect = "Allow"
#       Action = "secretsmanager:GetSecretValue"
#       Principal = {
#         AWS = aws_iam_role.github_actions.arn
#       }
#       Resource = "*"
#     }
#   })
# }
