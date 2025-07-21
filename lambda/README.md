# Lambda Extension Testing

This directory contains a testing framework for murmur's Lambda extension. It provides a local testing environment using AWS SAM to validate that the extension correctly exports secrets to Lambda functions.

## Purpose

The test framework validates that:
- The extension registers correctly with the Lambda Extensions API
- Secrets are exported to the configured file location
- The exported secrets file is readable by the Lambda function

## Requirements

- **AWS SAM CLI** - for local Lambda simulation (`pip install aws-sam-cli` or follow [official installation guide](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html))
- **Docker** - required by SAM local (usually included with Docker Desktop)
- **jq** - for JSON output parsing (`brew install jq` on macOS, `apt-get install jq` on Ubuntu)

## Usage

### Run the Test

From the project root directory:

```bash
make test-lambda
```

This will:
1. Build all packages using `release-dry-run` (skipping publication)
2. Run TAP-compliant tests that validate the Lambda extension
3. Check package exists, SAM invoke succeeds, and secrets are correct

### Manual Testing

You can also run the test manually:

```bash
# Ensure you're in the lambda/test directory
cd lambda/test

# Invoke the test function
sam local invoke TestFunction

# Or invoke with custom event data
sam local invoke TestFunction -e events/test.json
```

## Test Design

### SAM Template (`template.yaml`)

- **LayerVersion**: Uses parameterized ContentUri for flexibility  
- **Function**: Simple Python function that reads exported secrets
- **Environment Variables**: Uses `passthrough:` provider to avoid dependency on real secrets

### Test Function (`function.py`)

The test function:
1. Reads secrets from `MURMUR_EXPORT_FILE` (default: `/tmp/secrets.env`)
2. Parses the dotenv format (`KEY=VALUE` pairs)
3. Returns the secrets as a simple JSON object

### Expected Output

Successful test output (TAP format):
```
1..3
ok 1 - Extension package found
ok 2 - SAM local invoke successful  
ok 3 - Extension exported expected secrets
```

## Configuration

The test uses these environment variables (defined in `template.yaml`):

- `SECRET_ONE`, `SECRET_TWO`, `SECRET_THREE`: Test secret values using `passthrough:` provider
- `MURMUR_EXPORT_FILE`: Location of exported secrets file (`/tmp/secrets.env`)
- `MURMUR_EXPORT_FORMAT`: Export format (`dotenv`)
- `MURMUR_EXPORT_REFRESH_INTERVAL`: Set to `0s` to disable periodic refresh during testing

## Troubleshooting

### Extension Not Found

If you get errors about missing extension files:
```bash
# Build all packages (including extension)
make release-dry-run
```

### SAM CLI Issues

If `sam local invoke` fails:
- Ensure Docker is running
- Check that AWS SAM CLI is properly installed
- Try `sam local start-lambda` for debugging
