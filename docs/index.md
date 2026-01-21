---
page_title: "Vidos Provider"
description: "Interact with Vidos decentralized identity and access management services."
layout: provider
---

# Vidos Provider

The Vidos Terraform provider enables management of decentralized identity, access, and policy resources via the Vidos platform.

- [Vidos Product Documentation](https://vidos.id/docs)

## Example Usage

```hcl
provider "vidos" {
  region  = var.vidos_region
  api_key = var.vidos_api_key
}
```

## Authentication

- `api_key` (required): Your Vidos API key. Can also be set via the `VIDOS_API_KEY` environment variable.
- `region` (required): The Vidos region to use. Can also be set via the `VIDOS_REGION` environment variable.

## Environment Variables

- `VIDOS_API_KEY` – API key for authentication
- `VIDOS_REGION` – Region for resource operations

## Version and Domain Notes

- The provider defaults to the public Vidos cloud endpoints. For custom/private deployments, see [Vidos docs](https://vidos.id/docs).
- Provider versioning follows [Terraform Registry conventions](https://www.terraform.io/docs/registry/providers/publishing.html#versioning).

## Resources

See the [resources documentation](./resources/) for all supported resources.

For more details, visit the [Vidos documentation](https://vidos.id/docs).
