---
page_title: "vidos_authorizer_configuration Resource"
description: "Manage Vidos authorizer service configuration."
layout: resource
---

# vidos_authorizer_configuration

Manage authorizer service configuration in Vidos. Authorizer configurations define policies and settings for authorization decisions.

## Example Usage

```hcl
resource "vidos_authorizer_configuration" "example" {
  name = "terraform-example-authorizer-config"

  values = jsonencode({
    policies = {
      verify = {
        skip = true
      }
      validate = {
        skip = true
      }
    }
  })
}
```

## Argument Reference

- `name` (required) – Name of the authorizer configuration
- `values` (required) – JSON-encoded configuration values. See the [Vidos authorizer configuration documentation](https://vidos.id/docs/reference/services/authorizer/configuration/) for available configuration options.
- `resource_id` (optional) – Authorizer configuration resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the authorizer configuration (read-only if not provided)

## Import

Import an existing configuration by `resource_id`:

```bash
terraform import vidos_authorizer_configuration.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
