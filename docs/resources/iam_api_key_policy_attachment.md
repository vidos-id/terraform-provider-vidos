---
page_title: "vidos_iam_api_key_policy_attachment Resource"
description: "Attach an IAM policy to an API key in Vidos."
layout: resource
---

# vidos_iam_api_key_policy_attachment

Attach an IAM policy to an API key to grant scoped permissions.

## Example Usage

```hcl
resource "vidos_iam_api_key_policy_attachment" "example" {
  api_key_id  = vidos_iam_api_key.example.resource_id
  policy_type = "inline"
  policy_id   = vidos_iam_policy.example.resource_id
}
```

## Argument Reference

- `api_key_id` (required) – Resource ID of the API key
- `policy_type` (required) – Type of policy (e.g., "inline")
- `policy_id` (required) – Resource ID of the policy to attach

## Attributes Reference

- `id` – The attachment identifier (read-only)

## Import

Import an existing attachment:

```bash
terraform import vidos_iam_api_key_policy_attachment.example <attachment_id>
```

For more information, see the [Vidos IAM documentation](https://vidos.id/docs).
