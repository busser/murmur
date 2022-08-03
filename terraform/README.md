# Cloud Infrastucture <!-- omit in toc -->

This directory contains all Terraform code required to provision the cloud
resources used to test whisper functionality.

- [Requirements](#requirements)
- [Usage](#usage)
- [Documentation](#documentation)

## Requirements

Our recommended setup requires that you install [tfswitch](https://tfswitch.warrensbox.com/Install/).
Alternatively, you can install a version of the [Terraform CLI](https://www.terraform.io/downloads.html)
that matches the version constraints of this project.

You must be authenticated as a Microsoft Azure account that has the necessary
permissions to manage the infrastructure. To log in, run this command:

```bash
az login
```

You must also set the `GITHUB_TOKEN` environment variable to a personal access
token with the necessary permissions to manage this repository's secrets, and
the `GITHUB_OWNER` environment variable to `busser`, this repository's owner.

To generate documentation, you must install [terraform-docs](https://terraform-docs.io/)
and [prettier](https://prettier.io/).

## Usage

This repository contains the following layers:

- `layers/bootstrap`: resources that must exist before the other layers can be applied.
- `layers/azure-keyvault`: resources to test integration with Azure Key Vault.

To apply a layer, run the following commands in the layer's directory:

```bash
# Download the latest version of Terraform that matches the layer's version contraints.
tfswitch

# Initialise the layer.
terraform init

# Apply the layer.
terraform apply
```

## Documentation

Each directory with Terraform code in it has auto-generated documentation. To
update this documentation, run this command:

```bash
make docs
```
