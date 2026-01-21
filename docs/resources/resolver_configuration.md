---
page_title: "vidos_resolver_configuration Resource"
description: "Manage Vidos resolver service configuration."
layout: resource
---

# vidos_resolver_configuration

Manage resolver service configuration in Vidos. Resolver configurations define how the service processes credential resolution requests.

## Example Usage

```hcl
resource "vidos_resolver_configuration" "example" {
  name = "terraform-resolver-config"

  values = jsonencode({
    "methods" : {
      "cheqd" : {
        "enabled" : false
      }
    }
  })
}
```

## Argument Reference

- `name` (required) – Name of the resolver configuration
- `values` (required) – JSON-encoded configuration values. See the [Vidos resolver configuration documentation](https://vidos.id/docs/reference/services/resolver/configuration/) for available configuration options.
- `resource_id` (optional) – Resolver configuration resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the resolver configuration (read-only if not provided)

## Import

Import an existing configuration by `resource_id`:

```bash
terraform import vidos_resolver_configuration.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
