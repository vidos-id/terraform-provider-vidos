---
page_title: "vidos_iam_policy Resource"
description: "Manage a Vidos IAM policy for fine-grained access control."
layout: resource
---

# vidos_iam_policy

Manage a Vidos IAM policy that defines permissions and can be attached to API keys or service roles.

## Example Usage

```hcl
resource "vidos_iam_policy" "example" {
  name = "terraform-example-policy"
  
  document = jsonencode({
    version = "1.0"
    permissions = [
      {
        effect = "allow"
        scope  = "management"
        actions = ["read", "list"]
        resources = [
          {
            region       = "global"
            service      = "iam"
            resourceType = "*"
            resourceId   = "*"
          }
        ]
      }
    ]
  })
}
```

## Argument Reference

- `name` (required) – Name of the policy
- `document` (required) – JSON-encoded policy document defining permissions. See the [Vidos IAM policy documentation](https://vidos.id/docs/reference/services/gateway/configuration/) for schema and format details.

## Attributes Reference

- `resource_id` – Unique identifier for the policy (read-only)

## Import

Import an existing policy by `resource_id`:

```bash
terraform import vidos_iam_policy.example <resource_id>
```

For more information, see the [Vidos IAM documentation](https://vidos.id/docs).
