---
page_title: "vidos_validator_instance Resource"
description: "Manage a Vidos validator service instance."
layout: resource
---

# vidos_validator_instance

Manage a validator service instance in Vidos. An instance uses a validator configuration and provides the running validation service endpoint.

## Example Usage

```hcl
resource "vidos_validator_instance" "example" {
  name = "terraform-example-validator-instance"

  configuration_resource_id = vidos_validator_configuration.example.resource_id

  # Optional: provide inline configuration instead of a reference
  # inline_configuration = jsonencode({
  #   policies = {
  #     trustedIssuer = {
  #       skip = false
  #     }
  #   }
  # })
}
```

## Argument Reference

- `name` (required) – Name of the validator instance
- `configuration_resource_id` (optional) – Resource ID of a validator configuration to use
- `inline_configuration` (optional) – JSON-encoded inline configuration (alternative to configuration_resource_id). See the [Vidos validator configuration documentation](https://vidos.id/docs/reference/services/validator/configuration/) for available options.
- `resource_id` (optional) – Validator instance resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the validator instance (read-only if not provided)

## Import

Import an existing instance by `resource_id`:

```bash
terraform import vidos_validator_instance.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
