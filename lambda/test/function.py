import os


def lambda_handler(event, context):
    """
    Simple test function that reads exported secrets and returns them as JSON.
    """
    secrets_file = os.environ.get("MURMUR_EXPORT_FILE", "/tmp/secrets.env")

    secrets = {}
    with open(secrets_file, "r") as f:
        for line in f:
            line = line.strip()
            if line and "=" in line and not line.startswith("#"):
                key, value = line.split("=", 1)
                secrets[key] = value

    return secrets
