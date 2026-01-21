---
page_title: "vidos_iam_api_key Resource"
description: "Manage a Vidos IAM API key for authentication and authorization."
layout: resource
---

# vidos_iam_api_key

Manage a Vidos IAM API key. API keys are used for authentication and can have optional inline policies to scope permissions.

## Example Usage

```hcl
resource "vidos_iam_api_key" "example" {
  name = "terraform-example"

  # Optional: restrict permissions with an inline policy
  # inline_policy_document = jsonencode({
  #   version = "1.0"
  #   permissions = [
  #     {
  #       effect = "allow"
  #       scope  = "management"
  #       actions = ["read", "list"]
  #       resources = [
  #         {
  #           region       = "global"
  #           service      = "iam"
  #           resourceType = "*"
  #           resourceId   = "*"
  #         }
  #       ]
  #     }
  #   ]
  # })
}
```

## Argument Reference

- `name` (required) – Name of the API key
- `inline_policy_document` (optional) – JSON-encoded policy document to scope API key permissions. See the [Vidos IAM policy documentation](https://vidos.id/docs/reference/services/gateway/configuration/) for policy schema details.

## Attributes Reference

- `resource_id` – Unique identifier for the API key (read-only)
- `api_secret` – Secret associated with the API key. Sensitive (read-only)

## Import

Import an existing API key by `resource_id`:

```bash
terraform import vidos_iam_api_key.example <resource_id>
```

For more information, see the [Vidos IAM documentation](https://vidos.id/docs).
