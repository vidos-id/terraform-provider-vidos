# Examples (terraform-provider-vidos)

These examples are meant to be run **locally** against the provider in this repo.

## Prerequisites

- Terraform `>= 1.6`
- Go (same version you use to build the provider)

## Run locally (recommended)

Terraform won’t automatically find an unpublished provider, so use a Terraform CLI config with a `dev_overrides` entry.

1) Build the provider binary:

```bash
cd tools/terraform-provider-vidos

# Standard build (version defaults to 1.0.0)
go build -o ./bin/terraform-provider-vidos .

# Or inject a specific version at build time:
go build -ldflags "-X main.version=1.0.0" -o ./bin/terraform-provider-vidos .
```

2) Create a Terraform CLI config (example path: `~/.terraformrc`):

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/vidos-id/vidos" = "/ABSOLUTE/PATH/TO/verified-os/tools/terraform-provider-vidos/bin"
  }

  direct {}
}
```

3) Run an example.

### IAM example

```bash
cd tools/terraform-provider-vidos/examples/iam

export TF_CLI_CONFIG_FILE="$HOME/.terraformrc"
export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"
export TF_VAR_managed_policy_id="<EXISTING_MANAGED_POLICY_ID>"

terraform init
terraform apply
```

### Resolver example

```bash
cd tools/terraform-provider-vidos/examples/resolver

export TF_CLI_CONFIG_FILE="$HOME/.terraformrc"
export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

### Verifier example

```bash
cd tools/terraform-provider-vidos/examples/verifier

export TF_CLI_CONFIG_FILE="$HOME/.terraformrc"
export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

This example also creates a resolver configuration + instance, then points the verifier configuration at that resolver instance via `resolver.resourceId`.

### Validator example

```bash
cd tools/terraform-provider-vidos/examples/validator

export TF_CLI_CONFIG_FILE="$HOME/.terraformrc"
export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

### Authorizer example

```bash
cd tools/terraform-provider-vidos/examples/authorizer

export TF_CLI_CONFIG_FILE="$HOME/.terraformrc"
export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

### Gateway + Authorizer example

This example provisions both an authorizer and a gateway, and configures the gateway to route `/auth/*` to the authorizer instance using the managed service role `authorizer_all_actions`.

```bash
cd tools/terraform-provider-vidos/examples/gateway-authorizer

export TF_CLI_CONFIG_FILE="$HOME/.terraformrc"
export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

## Notes

- IAM uses the `global` management endpoint by design; `region` in the provider config is for non-IAM services (like resolver).
- `vidos_iam_api_key.api_secret` is write-only: it is returned only on create and cannot be recovered after import.
- Attachment resources fail fast by checking the policy exists before attaching.
