---
page_title: "Vidos Provider"
description: "Interact with Vidos decentralized identity and access management services."
layout: provider
---

# Vidos Provider

The Vidos Terraform provider enables management of decentralized identity, access, and policy resources via the Vidos platform.

This provider is officially published to the [Terraform Registry](https://registry.terraform.io/providers/vidos-id/vidos) and is available for use in your Terraform configurations.

- [Vidos Product Documentation](https://vidos.id/docs)
- [GitHub Repository](https://github.com/vidos-id/terraform-provider-vidos)

## Example Usage

To use this provider, declare it in your `terraform` configuration block:

```hcl
terraform {
  required_providers {
    vidos = {
      source  = "registry.terraform.io/vidos-id/vidos"
      version = "~> 0.1"
    }
  }
}

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

## Version Compatibility

This provider is compatible with:
- **Terraform**: 1.0 or later
- **Protocol Version**: 6.0 (Terraform Plugin Framework)

### Version Constraints

It is recommended to use version constraints in your configuration to ensure stability:

```hcl
terraform {
  required_providers {
    vidos = {
      source  = "registry.terraform.io/vidos-id/vidos"
      version = "~> 0.1"  # Allows 0.1.x updates
    }
  }
  required_version = ">= 1.0"
}
```

### Versioning

This provider follows [semantic versioning](https://semver.org/):
- **Major versions** (1.x.x, 2.x.x) may include breaking changes
- **Minor versions** (x.1.x, x.2.x) add new features in a backward-compatible manner
- **Patch versions** (x.x.1, x.x.2) include backward-compatible bug fixes

See the [CHANGELOG](https://github.com/vidos-id/terraform-provider-vidos/releases) for release notes and upgrade guides.

## Domain and Endpoint Configuration

- The provider defaults to the public Vidos cloud endpoints. For custom/private deployments, see [Vidos docs](https://vidos.id/docs).
- Provider versioning follows [Terraform Registry conventions](https://www.terraform.io/docs/registry/providers/publishing.html#versioning).

## Resources

See the [resources documentation](./resources/) for all supported resources.

## Guides

- [Guides index](./guides/)
- [Gateway + Authorizer](./guides/gateway-authorizer.md)
- [Gateway + Authorizer + Validator](./guides/gateway-authorizer-validator.md)

For more details, visit the [Vidos documentation](https://vidos.id/docs).
