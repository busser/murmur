# 🤫 Murmur <!-- omit in toc -->

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/busser/murmur)](https://goreportcard.com/report/github.com/busser/murmur)
![tests-passing](https://github.com/busser/murmur/actions/workflows/ci.yml/badge.svg)

Plug-and-play executable to pass secrets as environment variables to a process.

Murmur is a small binary that reads its environment variables, replaces
references to secrets with the secrets' values, and passes the resulting
variables to your application. Variables that do not reference secrets are
passed as-is.

Several tools like Murmur exist, each supporting a different secret provider.
Murmur aims to support as many providers as possible, so you can use Murmur no
matter which provider you use.

|                                                            | Scaleway | AWS | Azure | GCP | Vault | 1Password | Doppler |
| ---------------------------------------------------------- | -------- | --- | ----- | --- | ----- | --------- | ------- |
| 🤫 Murmur                                                  | ✅       | ✅  | ✅    | ✅  | ❌    | ❌        | ❌      |
| [Berglas](https://github.com/GoogleCloudPlatform/berglas)  | ❌       | ❌  | ❌    | ✅  | ❌    | ❌        | ❌      |
| [Bank Vaults](https://github.com/banzaicloud/bank-vaults)  | ❌       | ❌  | ❌    | ❌  | ✅    | ❌        | ❌      |
| [1Password CLI](https://developer.1password.com/docs/cli/) | ❌       | ❌  | ❌    | ❌  | ❌    | ✅        | ❌      |
| [Doppler CLI](https://github.com/DopplerHQ/cli)            | ❌       | ❌  | ❌    | ❌  | ❌    | ❌        | ✅      |

_If you know of a similar tool that is not listed here, please open an issue so
that we can add it to the list._

_If you use a secret provider that is not supported by Murmur, please open an
issue so that we can track demand for it._

- [Fetching a database password](#fetching-a-database-password)
- [Adding Murmur to a container image](#adding-murmur-to-a-container-image)
- [Adding Murmur to a Kubernetes pod](#adding-murmur-to-a-kubernetes-pod)
- [Parsing JSON secrets](#parsing-json-secrets)
- [Providers and filters](#providers-and-filters)
  - [`scwsm` provider: Scaleway Secret Manager](#scwsm-provider-scaleway-secret-manager)
  - [`awssm` provider: AWS Secrets Manager](#awssm-provider-aws-secrets-manager)
  - [`azkv` provider: Azure Key Vault](#azkv-provider-azure-key-vault)
  - [`gcpsm` provider: GCP Secret Manager](#gcpsm-provider-gcp-secret-manager)
  - [`passthrough` provider: no-op](#passthrough-provider-no-op)
  - [`jsonpath` filter: JSON parsing and templating](#jsonpath-filter-json-parsing-and-templating)
- [Changes from v0.4 to v0.5](#changes-from-v04-to-v05)

## Fetching a database password

Murmur runs as a wrapper around any command. For example, if you want to connect
to a PostgreSQL database, instead of running this command:

```bash
export PGPASSWORD="Q-gVzyDPmvsX6rRAPVjVjvfvR@KGzPJzCEg2"
psql -h 10.1.12.34 -U my-user -d my-database
```

You run this instead:

```bash
export PGPASSWORD="scwsm:database-password"
murmur run -- psql -h 10.1.12.34 -U my-user -d my-database
```

Murmur will fetch the value of the `database-password` secret from Scaleway
Secret Manager, set the `PGPASSWORD` environment variable to that value, and
then run `psql`.

## Adding Murmur to a container image

Murmur is a static binary, so you can simply copy it into your container image
and use it as your entrypoint. For convenience, the murmur binary is released as
a container image you can copy from in your Dockerfile:

```dockerfile
COPY --from=ghcr.io/busser/murmur:latest /murmur /bin/murmur
```

Then you can change your image's entrypoint:

```dockerfile
# from this:
ENTRYPOINT ["/bin/run-my-app"]
# to this:
ENTRYPOINT ["/bin/murmur", "run", "--", "/bin/run-my-app"]
```

## Adding Murmur to a Kubernetes pod

You can use Murmur in a Kubernetes pod even if your application's container
image does not include Murmur. To do so, you can use an init container that
copies Murmur into an emptyDir volume, and then use that volume in your
application's container.

Here is an example:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-app
spec:
  initContainers:
    - name: copy-murmur
      image: ghcr.io/busser/murmur:latest
      command: ["cp", "/murmur", "/shared/murmur"]
      volumeMounts:
        - name: shared
          mountPath: /shared
  containers:
    - name: my-app
      image: my-app:latest
      command: ["/shared/murmur", "run", "--", "/bin/run-my-app"]
      volumeMounts:
        - name: shared
          mountPath: /shared
  volumes:
    - name: shared
      emptyDir: {}
```

## Parsing JSON secrets

Storing secrets as JSON is a common pattern. For example, a secret might contain
a JSON object with multiple fields:

```json
{
  "host": "10.1.12.34",
  "port": 5432,
  "database": "my-database",
  "username": "my-user",
  "password": "Q-gVzyDPmvsX6rRAPVjVjvfvR@KGzPJzCEg2"
}
```

Murmur can parse that JSON and set environment variables for each field by using
the `jsonpath` filter:

```bash
export PGHOST="scwsm:database-credentials|jsonpath:{.host}"
export PGPORT="scwsm:database-credentials|jsonpath:{.port}"
export PGDATABASE="scwsm:database-credentials|jsonpath:{.database}"
export PGUSER="scwsm:database-credentials|jsonpath:{.username}"
export PGPASSWORD="scwsm:database-credentials|jsonpath:{.password}"
murmur run -- psql
```

If you have multiple references to the same secret, Murmur will fetch the secret
only once to avoid unnecessary API calls.

Alternatively, you can use the `jsonpath` filter to set a single environment
variable with the entire JSON object:

```bash
# psql supports connection strings, so we can use a single variable
export PGDATABASE="scwsm:database-credentials|jsonpath:postgres://{.username}:{password}@{.host}:{.port}/{.database}"
murmur run -- psql
```

Murmur uses the Kubernetes JSONPath syntax for the `jsonpath` filter. See the
[Kubernetes documentation](https://kubernetes.io/docs/reference/kubectl/jsonpath/)
for a full list of capabilities.

## Providers and filters

Murmur's architecture is built around providers and filters. Providers fetch
secrets from a secret manager, and filters parse and transform the secrets.

Murmur only edits environment variables which contain valid queries. A valid
query is structured as follows:

```plaintext
provider_id:secret_ref|filter_id:filter_rule
```

Using a filter is optional, so this is also a valid query:

```plaintext
provider_id:secret_ref
```

Murmur does not support chaining multiple filters yet.

### `scwsm` provider: Scaleway Secret Manager

To fetch a secret from [Scaleway Secret Manager](https://www.scaleway.com/en/secret-manager/),
the query must be structured as follows:

```plaintext
scwsm:[region/]{name|id}[#version]
```

If `region` is not specified, Murmur will delegate region selection to the
Scaleway SDK. The SDK determines the region based on the environment, by looking
at environment variables and configuration files.

One of `name` or `id` must be specified. Murmur guesses whether the string is a
name or an ID depending on whether it is a valid UUID. UUIDs are treated as IDs,
and other strings are treated as names.

The `version` must either be a positive integer or the "latest" string. If
`version` is not specified, Murmur defaults to "latest".

Examples:

```plaintext
scwsm:my-secret
scwsm:my-secret#123
scwsm:my-secret#latest

scwsm:fr-par/my-secret
scwsm:fr-par/my-secret#123
scwsm:fr-par/my-secret#latest

scwsm:3f34b83f-47a6-4344-bcd4-b63721481cd3
scwsm:3f34b83f-47a6-4344-bcd4-b63721481cd3#123
scwsm:3f34b83f-47a6-4344-bcd4-b63721481cd3#latest

scwsm:fr-par/3f34b83f-47a6-4344-bcd4-b63721481cd3
scwsm:fr-par/3f34b83f-47a6-4344-bcd4-b63721481cd3#123
scwsm:fr-par/3f34b83f-47a6-4344-bcd4-b63721481cd3#latest
```

Murmur uses the environment's default credentials to authenticate to Scaleway.
You can configure Murmur the same way you can [configure the `scw` CLI](https://github.com/scaleway/scaleway-cli/blob/master/docs/commands/config.md).

### `awssm` provider: AWS Secrets Manager

To fetch a secret from [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/),
the query must be structured as follows:

```plaintext
awssm:{name|arn}[#{version_id|version_stage}]
```

One of `name` or `arn` must be specified. You can use a full or partial ARN.
However, if your secret's name ends with a hyphen followed by six characters,
you should not use a partial ARN. See [these AWS docs](https://docs.aws.amazon.com/secretsmanager/latest/userguide/troubleshoot.html#ARN_secretnamehyphen)
for more information.

You can optionally specify one of `version_id` or `version_stage`. Murmur
guesses whether the string is an ID or a stage depending on whether it is a
valid UUID. UUIDs are treated as version IDs, and other strings are treated as
version stages. If neither `version_id` or `version_stage` are specified, Murmur
defaults to "AWSCURRENT".

Examples:

```plaintext
awssm:my-secret
awssm:my-secret#MY_VERSION_STAGE
awssm:my-secret#9517cc59-646a-4393-81d7-5e6f2d43cbe7

awssm:arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret
awssm:arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret#MY_VERSION_STAGE
awssm:arn:aws:secretsmanager:us-east-1:123456789012:secret:my-secret#9517cc59-646a-4393-81d7-5e6f2d43cbe7
```

Murmur uses the environment's default credentials to authenticate to AWS.
You can configure Murmur the same way you can [configure the `aws` CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html).

### `azkv` provider: Azure Key Vault

To fetch a secret from [Azure Key Vault](https://azure.microsoft.com/en-us/services/key-vault/),
the query must be structured as follows:

```plaintext
azkv:keyvault_hostname/name[#version]
```

The `keyvault_hostname` must be the fully qualified domain name of the Key
Vault. For example, if your Key Vault's URL is `https://example.vault.azure.net/`,
then the `keyvault_hostname` is `example.vault.azure.net`.

The `name` is the name of the secret.

The `version` must be a valid version ID. If `version` is not specified, Murmur
defaults to the latest version of the secret.

Examples:

```plaintext
azkv:example.vault.azure.net/my-secret
azkv:example.vault.azure.net/my-secret#5ddc29704c1c4429a4c53605b7949100
```

Murmur uses the environment's default credentials to authenticate to Azure. You
can set these credentials with the [environment variables listed here](https://github.com/Azure/azure-sdk-for-go/wiki/Set-up-Your-Environment-for-Authentication#configure-defaultazurecredential),
or with workload identity.

### `gcpsm` provider: GCP Secret Manager

To fetch a secret from [GCP Secret Manager](https://cloud.google.com/secret-manager),
the query must be structured as follows:

```plaintext
gcpsm:project/name[#version]
```

The `project` must be either a project ID or a project number.

The `name` is the name of the secret.

The `version` must be a valid version number. If `version` is not specified,
Murmur defaults to the latest version of the secret.

### `passthrough` provider: no-op

This provider is meant for demo and testing purposes. It does not fetch any
secrets and simply returns the secret reference as the secret's value.

This provider, like all other providers, is fully tested. It is safe to use in
production, although why would you?

Examples:

```plaintext
passthrough:my-not-so-secret-value
```

### `jsonpath` filter: JSON parsing and templating

To parse a JSON secret and extract a value from it, or to use a secret value in
a template, the query must be stuctured as follows:

```plaintext
provider_id:secret_ref|jsonpath:template
```

The `provider_id` and `secret_ref` can be any valid secret reference.

The `template` is a [JSONPath template](https://kubernetes.io/docs/reference/kubectl/jsonpath/).
Murmur uses the Kubernetes JSONPath implementation, so you can use any feature
described in the Kubernetes docs.

If the secret's value is not valid JSON, Murmur will treat it as a string and
execute the template anyway. This means that you can use JSONPath templates with
non-JSON secrets.

Examples:

```plaintext
scwsm:my-secret|jsonpath:{.password}
scwsm:my-secret|jsonpath:postgres://{.username}:{.password}@{.hostname}:{.port}/{.database}
scwsm:my-secret|jsonpath:the secret is {@}
```

## Changes from v0.4 to v0.5

Following community feedback, we have made two significant changes in v0.5:

1. We have renamed the project from "Whisper" to "Murmur", to make the project
   documentation easier to find on search engines.
2. We have renamed the `exec` command to `run`, to make it clear that we are not
   executing the command directly, but rather running it as a subprocess.

We have made it so that none of these changes are breaking. You can upgrade to
v0.5 without changing anything in how you use Whisper/Murmur.

We now publish binaries and container images with both names. The `exec` command
is still available, but it will log a warning message telling you to use the new
`run` command instead.

We recommend that you update your scripts to use the new name and command, but
you have all the time you need to do so.
