---
page_title: "vidos_verifier_configuration Resource"
description: "Manage Vidos verifier service configuration."
layout: resource
---

# vidos_verifier_configuration

Manage verifier service configuration in Vidos. Verifier configurations define how the service processes verification requests and policies.

## Example Usage

```hcl
resource "vidos_verifier_configuration" "example" {
  name = "terraform-example-verifier-config"

  values = jsonencode({
    cors = {
      enabled = true
      origin  = ["*"]
    }

    policies = {
      proof = {
        skip = false
      }
    }
  })
}
```

## Argument Reference

- `name` (required) – Name of the verifier configuration
- `values` (required) – JSON-encoded configuration values. See the [Vidos verifier configuration documentation](https://vidos.id/docs/reference/services/verifier/configuration/) for available configuration options.
- `resource_id` (optional) – Verifier configuration resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the verifier configuration (read-only if not provided)

## Import

Import an existing configuration by `resource_id`:

```bash
terraform import vidos_verifier_configuration.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
