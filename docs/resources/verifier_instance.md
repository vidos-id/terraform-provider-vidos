---
page_title: "vidos_verifier_instance Resource"
description: "Manage a Vidos verifier service instance."
layout: resource
---

# vidos_verifier_instance

Manage a verifier service instance in Vidos. An instance uses a verifier configuration and provides the running verification service endpoint.

## Example Usage

```hcl
resource "vidos_verifier_instance" "example" {
  name = "terraform-example-verifier-instance"

  configuration_resource_id = vidos_verifier_configuration.example.resource_id

  # Optional: provide inline configuration instead of a reference
  # inline_configuration = jsonencode({
  #   cors = {
  #     enabled = true
  #     origin  = ["*"]
  #   }
  # })
}
```

## Argument Reference

- `name` (required) – Name of the verifier instance
- `configuration_resource_id` (optional) – Resource ID of a verifier configuration to use
- `inline_configuration` (optional) – JSON-encoded inline configuration (alternative to configuration_resource_id). See the [Vidos verifier configuration documentation](https://vidos.id/docs/reference/services/verifier/configuration/) for available options.
- `resource_id` (optional) – Verifier instance resource ID. Immutable. If omitted, the provider will generate one.

## Attributes Reference

- `resource_id` – Unique identifier for the verifier instance (read-only if not provided)

## Import

Import an existing instance by `resource_id`:

```bash
terraform import vidos_verifier_instance.example <resource_id>
```

For more information, see the [Vidos documentation](https://vidos.id/docs).
