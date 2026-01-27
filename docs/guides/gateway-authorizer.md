---
page_title: "Gateway + Authorizer Example"
description: "Provision an authorizer instance and a gateway that routes /auth/* to it."
layout: guide
---

# Gateway + Authorizer

This guide explains the `examples/gateway-authorizer` configuration.

## What it provisions

- `vidos_authorizer_configuration` (empty config scaffold)
- `vidos_authorizer_instance`
- `vidos_gateway_instance` with an inline configuration that routes `/<path>` requests to backing services

In this example, the gateway forwards requests under `/auth/*` to the authorizer instance, using the managed service role `authorizer_all_actions` for gateway -> authorizer service-to-service authorization.

## Prerequisites

- Terraform `>= 1.6`
- A Vidos IAM API secret with permissions to create authorizer and gateway resources

## Inputs

- `vidos_api_key` (required, sensitive): IAM API secret
- `vidos_region` (optional, default `eu`): service region

You can provide these as Terraform variables or via environment variables:

- `TF_VAR_vidos_api_key` (maps to `vidos_api_key`)
- `VIDOS_API_KEY` (provider auth)
- `VIDOS_REGION` (provider region)

## Run it

```bash
cd examples/gateway-authorizer

export TF_VAR_vidos_api_key="<YOUR_VIDOS_IAM_API_SECRET>"

terraform init
terraform apply
```

Clean up:

```bash
terraform destroy
```

## How the routing works

The gateway instance uses `inline_configuration` with a `paths` map. The `auth` entry configures `/auth/*` to be forwarded to the authorizer instance:

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
}
```

## Customize

- Restrict CORS: replace `origin = ["*"]` and `allowHeaders = ["*"]` with your expected domains/headers.
- Add more routes: add more keys under `paths` (each key becomes a `/<key>/*` prefix).
- Use a reusable configuration: create a `vidos_gateway_configuration` and set `configuration_resource_id` instead of `inline_configuration`.
