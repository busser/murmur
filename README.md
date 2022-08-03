# 🤫 Whisper <!-- omit in toc -->

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/busser/whisper)](https://goreportcard.com/report/github.com/busser/whisper)
![tests-passing](https://github.com/busser/whisper/actions/workflows/ci.yml/badge.svg)
![go-version](https://img.shields.io/badge/Go-v1.18-29beb0)

Plug-and-play entrypoint to inject secrets directly into your application's
environment variables.

- [How it works](#how-it-works)
- [Using whisper locally](#using-whisper-locally)
- [Including whisper in a Docker image](#including-whisper-in-a-docker-image)
- [Secret providers](#secret-providers)
  - [Azure Key Vault](#azure-key-vault)
  - [Google Secret Manager](#google-secret-manager)
  - [Hashicorp Vault](#hashicorp-vault)
  - [AWS Secrets Manager](#aws-secrets-manager)
  - [Passthrough](#passthrough)
- [Troubleshooting](#troubleshooting)
  - [`Error: unknown flag`](#error-unknown-flag)

## How it works

Whisper must run as your application's entrypoint. This means that instead of
running this command to start your application:

```bash
/bin/run-my-app
```

Run this instead:

```bash
whisper exec -- /bin/run-my-app
```

Whisper reads its environment variables, replaces references to secrets with
the secrets' values, and passes the resulting variables to your application.
Variables that are not references to secrets are passed as is.

## Using whisper locally

Download the `whisper` binary for your OS and architecture on the
[project's releases page](https://github.com/busser/whisper/releases) and put
the binary in your PATH.

## Including whisper in a Docker image

For convenience, the whisper binary is also released as a Docker image. In your
application's Dockerfile, simply add the following line:

```dockerfile
COPY --from=ghcr.io/busser/whisper:latest /whisper /bin/whisper
```

And then change your image's entrypoint:

```dockerfile
# from this:
ENTRYPOINT ["/bin/run-my-app"]
# to this:
ENTRYPOINT ["/bin/whisper", "exec", "--", "/bin/run-my-app"]
```

See [examples/dockerfile](./examples/dockerfile) for actual code.

## Secret providers

Whisper supports fetching secrets from the following providers.

### Azure Key Vault

Whisper will fetch secrets from Azure Key Vault for all environment variables
that start with `azkv:`. What follows the prefix should reference a secret.

Here are some examples:

- `azkv:example.vault.azure.net/secret-sauce` references the latest value of the
  `secret-sauce` secret in the `example` Key Vault.
- `azkv:example.vault.azure.net/secret-sauce/5ddc29704c1c4429a4c53605b7949100`
  references a specific version of the `secret-sauce` secret in the `example`
  Key Vault.

Whisper uses the environment's default credentials to authenticate to Azure. You
can set these credentials with the [environment variables listed here](https://github.com/Azure/azure-sdk-for-go/wiki/Set-up-Your-Environment-for-Authentication#configure-defaultazurecredential).

### Google Secret Manager

Not yet supported.

You may want to have a look at [berglas](https://github.com/GoogleCloudPlatform/berglas)
in the mean time.

### Hashicorp Vault

Not yet supported.

You mat want to have a look at [bank-vaults](https://github.com/banzaicloud/bank-vaults)
in the mean time.

### AWS Secrets Manager

Not yet supported.

### Passthrough

The `passthrough:` prefix is special: it does not fetch secrets from anywhere.
Whisper uses the secret's reference as its value. In effect, this simply removes
the `passthrough:` prefix from any environment variables.

## Troubleshooting

### `Error: unknown flag`

Your application may use flags, like this:

```bash
whisper exec /bin/run-my-app --port=3000
```

Whisper then picks up the `--port` flag and returns an error:

```plaintext
Error: unknown flag: --port
```

Whisper ignores any flags that come after a special `--` argument. So simply run
this command instead:

```bash
whisper exec -- /bin/run-my-app --port=3000
```

Any flags after the `--` argument will still be passed to your application.
