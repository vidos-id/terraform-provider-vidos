---
page_title: "vidos_resolver_instance Resource"
description: "Manage a Vidos resolver service instance."
layout: resource
---

# vidos_resolver_instance

Manage a resolver service instance in Vidos. An instance uses a resolver configuration and provides the running service endpoint.

## Example Usage

```hcl
resource "vidos_resolver_instance" "example" {
  name = "terraform-resolver-instance"

  configuration_resource_id = vidos_resolver_configuration.example.resource_id

  # Optional: provide inline configuration instead of a reference
  # inline_configuration = jsonencode({
  #   "methods" : {
  #     "cheqd" : {
  #       "enabled" : true
  #     }
  #   }
  # })
}
```

## Output Example

```hcl
output "resolver_endpoint" {
  value = vidos_resolver_instance.example.endpoint
}
```

## Argument Reference

- `name` (required) – Name of the resolver instance
- `configuration_resource_id` (optional) – Resource ID of a resolver configuration to use
- `inline_configuration` (optional) – JSON-encoded inline configuration (alternative to configuration_resource_id). See the [Vidos resolver configuration documentation](https://vidos.id/docs/reference/services/resolver/configuration/) for available options.
- `resource_id` (optional) – Resolver instance resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the resolver instance (read-only if not provided)
- `endpoint` – Platform-reported resolver endpoint (read-only)

## Import

Import an existing instance by `resource_id`:

```bash
terraform import vidos_resolver_instance.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
