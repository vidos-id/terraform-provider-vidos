---
page_title: "vidos_validator_configuration Resource"
description: "Manage Vidos validator service configuration."
layout: resource
---

# vidos_validator_configuration

Manage validator service configuration in Vidos. Validator configurations define policies and settings for credential validation.

## Example Usage

```hcl
resource "vidos_validator_configuration" "example" {
  name = "terraform-example-validator-config"

  values = jsonencode({
    policies = {
      trustedIssuer = {
        skip = true
      }
    }
  })
}
```

## Argument Reference

- `name` (required) – Name of the validator configuration
- `values` (required) – JSON-encoded configuration values. See the [Vidos validator configuration documentation](https://vidos.id/docs/reference/services/validator/configuration/) for available configuration options.
- `resource_id` (optional) – Validator configuration resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the validator configuration (read-only if not provided)

## Import

Import an existing configuration by `resource_id`:

```bash
terraform import vidos_validator_configuration.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
