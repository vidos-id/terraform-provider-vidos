---
page_title: "vidos_authorizer_instance Resource"
description: "Manage a Vidos authorizer service instance."
layout: resource
---

# vidos_authorizer_instance

Manage an authorizer service instance in Vidos. An instance uses an authorizer configuration and provides the running authorization service endpoint.

## Example Usage

```hcl
resource "vidos_authorizer_instance" "example" {
  name = "terraform-example-authorizer-instance"

  configuration_resource_id = vidos_authorizer_configuration.example.resource_id

  # Optional: provide inline configuration instead of a reference
  # inline_configuration = jsonencode({
  #   policies = {
  #     verify = {
  #       skip = false
  #     }
  #   }
  # })
}
```

## Argument Reference

- `name` (required) – Name of the authorizer instance
- `configuration_resource_id` (optional) – Resource ID of an authorizer configuration to use
- `inline_configuration` (optional) – JSON-encoded inline configuration (alternative to configuration_resource_id). See the [Vidos authorizer configuration documentation](https://vidos.id/docs/reference/services/authorizer/configuration/) for available options.
- `resource_id` (optional) – Authorizer instance resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the authorizer instance (read-only if not provided)

## Import

Import an existing instance by `resource_id`:

```bash
terraform import vidos_authorizer_instance.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
