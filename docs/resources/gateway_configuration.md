---
page_title: "vidos_gateway_configuration Resource"
description: "Manage Vidos gateway service configuration."
layout: resource
---

# vidos_gateway_configuration

Manage gateway service configuration in Vidos. Gateway configurations define routing and service integration settings.

## Example Usage

```hcl
resource "vidos_gateway_configuration" "example" {
  name   = "terraform-example-gateway-config"
  values = jsonencode({})
}
```

## Argument Reference

- `name` (required) – Name of the gateway configuration
- `values` (required) – JSON-encoded configuration values. See the [Vidos gateway configuration documentation](https://vidos.id/docs/reference/services/gateway/configuration/) for available configuration options.
- `resource_id` (optional) – Gateway configuration resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the gateway configuration (read-only if not provided)

## Import

Import an existing configuration by `resource_id`:

```bash
terraform import vidos_gateway_configuration.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
