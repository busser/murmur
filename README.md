# 🤫 Whisper <!-- omit in toc -->

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/busser/whisper)](https://goreportcard.com/report/github.com/busser/whisper)
![tests-passing](https://github.com/busser/whisper/actions/workflows/ci.yml/badge.svg)

Plug-and-play entrypoint to inject secrets directly into your application's
environment variables.

- [How it works](#how-it-works)
- [Using whisper locally](#using-whisper-locally)
- [Including whisper in a Docker image](#including-whisper-in-a-docker-image)
- [Secret providers](#secret-providers)
  - [Scaleway Secret Manager](#scaleway-secret-manager)
  - [Azure Key Vault](#azure-key-vault)
  - [AWS Secrets Manager](#aws-secrets-manager)
  - [Google Secret Manager](#google-secret-manager)
  - [Hashicorp Vault](#hashicorp-vault)
  - [Passthrough](#passthrough)
- [Filters](#filters)
  - [JSONPath](#jsonpath)
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
Variables that are not references to secrets are passed as is. See
[Secret providers](#secret-providers) below for more details.

Environment variable values can also contain filters that transform the secret's
value. See [Filters](#filters) below for more details.

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

### Scaleway Secret Manager

Whisper will fetch secrets from Scaleway Secret Manager for all environment
variables that start with `scwsm:`. What follows the prefix should reference a
secret.

Here are some examples:

- `scwsm:secret-sauce` references the latest value of the secret in the default
  region named `secret-sauce`.
- `scwsm:fr-par/secret-sauce` references the latest value of the secret in the
  `fr-par` region named `secret-sauce`.
- `scwsm:fr-par/3f34b83f-47a6-4344-bcd4-b63721481cd3` references the latest value
  of the secret in the `fr-par` region ID'd by
  `3f34b83f-47a6-4344-bcd4-b63721481cd3`.
- `scwsm:secret-sauce#123` references version `123` of the secret in the default
  region named `secret-sauce`.
- `scwsm:fr-par/secret-sauce#123` references version `123` of the secret in the
  `fr-par` region named `secret-sauce`.

The string that comes before `#` could be a name or an ID. If the string is a
UUID, then whisper assumes it is an ID. Otherwise, it assumes it is a name.

Whisper uses the environment's default credentials to authenticate to Scaleway.
You can configure whisper the same way you can [configure the `scw` CLI](https://github.com/scaleway/scaleway-cli/blob/master/docs/commands/config.md).

### Azure Key Vault

Whisper will fetch secrets from Azure Key Vault for all environment variables
that start with `azkv:`. What follows the prefix should reference a secret.

Here are some examples:

- `azkv:example.vault.azure.net/secret-sauce` references the latest value of the
  `secret-sauce` secret in the `example` Key Vault.
- `azkv:example.vault.azure.net/secret-sauce#5ddc29704c1c4429a4c53605b7949100`
  references a specific version of the `secret-sauce` secret in the `example`
  Key Vault.

Whisper uses the environment's default credentials to authenticate to Azure. You
can set these credentials with the [environment variables listed here](https://github.com/Azure/azure-sdk-for-go/wiki/Set-up-Your-Environment-for-Authentication#configure-defaultazurecredential),
or with workload identity.

### AWS Secrets Manager

Whisper will fetch secrets from AWS Secrets Manager for all environment
variables that start with `awssm:`. What follows the prefix should reference a
secret.

Here are some examples:

- `awssm:secret-sauce` references the current value of the `secret-sauce` secret
  in the region and account defined by the environment.
- `awssm:secret-sauce#9517cc59-646a-4393-81d7-5e6f2d43cbe7` references a
  specific version of the `secret-sauce` secret in the region and account
  defined by the environment.
- `awssm:secret-sauce#my-label` references a specific staging label of the
  `secret-sauce` secret in the region and account defined by the environment.
- `awssm:arn:aws:secretsmanager:us-east-1:123456789012:secret:secret-sauce-abcdef`
  references the secret with the specified ARN.
- `awssm:arn:aws:secretsmanager:us-east-1:123456789012:secret:secret-sauce-abcdef#my-label`
  references a specific staging label of the secret with the specified ARN.

The string that comes after `#` could be a version ID or a version label. If the
string is a UUID, then whisper assumes it is a version ID. Otherwise, it assumes
it is a version label.

Whisper uses the environment's default credentials to authenticate to AWS.

### Google Secret Manager

Whisper will fetch secrets from Google Cloud Platform's Secret Manager for all
environment variables that start with `gcpsm:`. What follows the prefix should
reference a secret.

Here are some examples:

- `gcpsm:example/secret-sauce` references the latest value of the
  `secret-sauce` secret in the `example` project.
- `gcpsm:example/secret-sauce#123` references a specific version of the
- `secret-sauce` secret in the `example` project.

Whisper uses the environment's default credentials to authenticate to Google
Cloud. You can set these with the `gcloud` CLI, with environment variables,
with Google Cloud's environment service accounts, or with workload identity.

An alternative to whisper, specific to Google Cloud, is [berglas](https://github.com/GoogleCloudPlatform/berglas).

### Hashicorp Vault

Not yet supported.

You mat want to have a look at [bank-vaults](https://github.com/banzaicloud/bank-vaults)
in the mean time.

### Passthrough

The `passthrough:` prefix is special: it does not fetch secrets from anywhere.
Whisper uses the secret's reference as its value. In effect, this simply removes
the `passthrough:` prefix from any environment variables.

## Filters

Whisper supports transforming secrets with the following filters.

### JSONPath

Whisper embeds the [Kubernetes JSONPath](https://kubernetes.io/docs/reference/kubectl/jsonpath/)
library. You can use it to extract specific fields from a JSON-encoded secret.
For example, if you have a secret with a value of `{"sauce": "szechuan"}`, the
`jsonpath` filter can extract the `sauce` field's value:

```plaintext
awssm:secret-sauce|jsonpath:{.sauce}
```

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
