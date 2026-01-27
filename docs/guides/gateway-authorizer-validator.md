---
page_title: "Gateway + Authorizer + Validator Example"
description: "Provision a validator, an authorizer configured to validate via that validator, and a gateway that routes to both."
layout: guide
---

# Gateway + Authorizer + Validator

This guide explains the `examples/gateway-authorizer-validator` configuration.

## What it provisions

- `vidos_validator_instance` with an inline policy configuration (trusted issuer roots)
- `vidos_authorizer_instance` configured to call the validator instance for validation
- `vidos_gateway_instance` that routes:
  - `/auth/*` to the authorizer
  - `/validate/*` directly to the validator

Managed service roles are used for service-to-service authorization:

- `authorizer_all_actions` for gateway -> authorizer
- `validator_all_actions` for gateway -> validator and authorizer -> validator

## Prerequisites

- Terraform `>= 1.6`
- A Vidos IAM API secret with permissions to create validator, authorizer, and gateway resources

## Inputs

- `vidos_api_key` (required, sensitive): IAM API secret
- `vidos_region` (optional, default `eu`): service region

## Run it

```bash
cd examples/gateway-authorizer-validator

export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

Clean up:

```bash
terraform destroy
```

## Validator trust anchors (important)

The example includes PEM-encoded root certificates under `locals` and configures them as trust anchors for issuer validation.

- Replace the example PEMs with your real roots.
- You can also include predefined roots (the example includes `{ type = "predefined", tag = "vidos" }`).

## How validation is wired

The authorizer is configured to validate using a validator *instance* reference:

```hcl
policies = {
  validate = {
    skip = false
    validator = {
      type       = "instance"
      resourceId = vidos_validator_instance.main.resource_id
      serviceRole = {
        owner      = "managed"
        resourceId = "validator_all_actions"
      }
    }
  }
}
```

## Gateway routing

The gateway defines two routes so you can hit validation directly (without going through the authorizer) when desired:

```hcl
paths = {
  auth = {
    type       = "instance"
    service    = "authorizer"
    resourceId = vidos_authorizer_instance.main.resource_id
    serviceRole = {
      owner      = "managed"
      resourceId = "authorizer_all_actions"
    }
  }

  validate = {
    type       = "instance"
    service    = "validator"
    resourceId = vidos_validator_instance.main.resource_id
    serviceRole = {
      owner      = "managed"
      resourceId = "validator_all_actions"
    }
  }
}
```
