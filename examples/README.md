# Examples (terraform-provider-vidos)

These examples are intended for **end users** of the published Terraform provider.

If you’re developing the provider from source and want to run the examples against a local binary, see **Run against a local build (contributors)** below.

## Prerequisites

- Terraform `>= 1.6`
- A Vidos API key with the permissions required by the example

## Run with the published provider (recommended)

Each example directory is a standalone Terraform configuration that downloads the provider from the Terraform Registry.

1) Pick an example and set required variables.

### Gateway + Authorizer example

This example provisions both an authorizer and a gateway, and configures the gateway to route `/auth/*` to the authorizer instance using the managed service role `authorizer_all_actions`.

Guide: `docs/guides/gateway-authorizer.md`

```bash
cd examples/gateway-authorizer

export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

2) Clean up when you’re done:

```bash
terraform destroy
```

### Gateway + Authorizer + Validator example

This example provisions a validator (with inline trusted issuer roots including example PEMs), an authorizer configured to use that validator, and a gateway that routes `/auth/*` to the authorizer.

Guide: `docs/guides/gateway-authorizer-validator.md`

```bash
cd examples/gateway-authorizer-validator

export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

## Notes

- IAM uses the `global` management endpoint by design; `region` in the provider config is for non-IAM services (like resolver).
- `vidos_iam_api_key.api_secret` is write-only: it is returned only on create and cannot be recovered after import.
- Attachment resources fail fast by checking the policy exists before attaching.
