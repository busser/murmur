# Bootstrap layer

This layer creates Scaleway buckets where the other layers' state will be stored.

> ⚠️ WARNING ⚠️
>
> The project should only need to be bootstrapped once. If you don't now what
> bootstrapping means, don't do it.

Because this layer's state contains no sensitive information, we store this
state in this repository.

> ⚠️ WARNING ⚠️
>
> Make sure that no sensitive information is ever stored in this repository's
> state.

<!-- BEGIN_TF_DOCS -->

## Requirements

| Name                                                                     | Version    |
| ------------------------------------------------------------------------ | ---------- |
| <a name="requirement_terraform"></a> [terraform](#requirement_terraform) | = 1.1.9    |
| <a name="requirement_scaleway"></a> [scaleway](#requirement_scaleway)    | 2.2.1-rc.3 |

## Providers

| Name                                                            | Version    |
| --------------------------------------------------------------- | ---------- |
| <a name="provider_scaleway"></a> [scaleway](#provider_scaleway) | 2.2.1-rc.3 |

## Modules

No modules.

## Resources

| Name                                                                                                                                        | Type     |
| ------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| [scaleway_object_bucket.terraform_state](https://registry.terraform.io/providers/scaleway/scaleway/2.2.1-rc.3/docs/resources/object_bucket) | resource |

## Inputs

No inputs.

## Outputs

No outputs.

<!-- END_TF_DOCS -->
