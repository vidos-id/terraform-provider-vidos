---
page_title: "vidos_iam_service_role_policy_attachment Resource"
description: "Attach an IAM policy to a service role in Vidos."
layout: resource
---

# vidos_iam_service_role_policy_attachment

Attach an IAM policy to a service role to grant scoped permissions for delegated access scenarios.

## Example Usage

```hcl
resource "vidos_iam_service_role_policy_attachment" "example" {
  service_role_id = vidos_iam_service_role.example.resource_id
  policy_type     = "inline"
  policy_id       = vidos_iam_policy.example.resource_id
}
```

## Argument Reference

- `service_role_id` (required) – Resource ID of the service role
- `policy_type` (required) – Type of policy (e.g., "inline")
- `policy_id` (required) – Resource ID of the policy to attach

## Attributes Reference

- `id` – The attachment identifier (read-only)

## Import

Import an existing attachment:

```bash
terraform import vidos_iam_service_role_policy_attachment.example <attachment_id>
```

For more information, see the [Vidos IAM documentation](https://vidos.id/docs).
