---
page_title: "vidos_iam_service_role Resource"
description: "Manage a Vidos IAM service role for delegated access."
layout: resource
---

# vidos_iam_service_role

Manage a Vidos IAM service role that allows services to assume permissions via role chaining.

## Example Usage

```hcl
resource "vidos_iam_service_role" "example" {
  name = "terraform-example-service-role"

  # Optional: define inline permissions
  # inline_policy_document = jsonencode({
  #   version = "1.0"
  #   permissions = [
  #     {
  #       effect  = "allow"
  #       scope   = "management"
  #       actions = ["read"]
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

- `name` (required) – Name of the service role
- `inline_policy_document` (optional) – JSON-encoded policy document for the role. See the [Vidos IAM policy documentation](https://vidos.id/docs/reference/services/gateway/configuration/) for policy schema details.

## Attributes Reference

- `resource_id` – Unique identifier for the service role (read-only)

## Import

Import an existing service role by `resource_id`:

```bash
terraform import vidos_iam_service_role.example <resource_id>
```

For more information, see the [Vidos IAM documentation](https://vidos.id/docs).
