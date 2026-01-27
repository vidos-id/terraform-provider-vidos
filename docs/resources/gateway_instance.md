---
page_title: "vidos_gateway_instance Resource"
description: "Manage a Vidos gateway service instance."
layout: resource
---

# vidos_gateway_instance

Manage a gateway service instance in Vidos. An instance uses a gateway configuration and provides the API gateway endpoint for service integration.

## Example Usage

```hcl
resource "vidos_gateway_instance" "example" {
  name = "terraform-example-gateway-instance"

  configuration_resource_id = vidos_gateway_configuration.example.resource_id

  # Optional: provide inline configuration instead of a reference
  # inline_configuration = jsonencode({})
}
```

## Output Example

```hcl
output "gateway_endpoint" {
  value = vidos_gateway_instance.example.endpoint
}
```

## Argument Reference

- `name` (required) – Name of the gateway instance
- `configuration_resource_id` (optional) – Resource ID of a gateway configuration to use
- `inline_configuration` (optional) – JSON-encoded inline configuration (alternative to configuration_resource_id). See the [Vidos gateway configuration documentation](https://vidos.id/docs/reference/services/gateway/configuration/) for available options.
- `resource_id` (optional) – Gateway instance resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the gateway instance (read-only if not provided)
- `endpoint` – Platform-reported gateway endpoint (read-only)

## Import

Import an existing instance by `resource_id`:

```bash
terraform import vidos_gateway_instance.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
