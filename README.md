# Terraform Provider: vidos

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/vidos-id/vidos/latest)

This is a Terraform provider for administering the Vidos platform via **management APIs only**.

## Installation

This provider is published to the [Terraform Registry](https://registry.terraform.io/providers/vidos-id/vidos/latest).

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    vidos = {
      source  = "registry.terraform.io/vidos-id/vidos"
      version = "~> 0.1"
    }
  }
}
```

### Requirements

- Terraform 1.0+
- Go 1.21+ (for local development)

## Scope (v1)

- Auth via an existing IAM API secret (`Authorization: Bearer <api_secret>`). No Cognito tokens.
- IAM:
  - Manage API keys.
  - Manage `account` policies.
  - Manage attachments from API keys to policies (policy may be `account` or `managed`).
  - Manage service roles.
  - Manage attachments from service roles to policies (policy may be `account` or `managed`).
- Authorizer:
  - Manage configurations.
  - Manage instances.
- Resolver:
  - Manage configurations.
  - Manage instances.
- Validator:
  - Manage configurations.
  - Manage instances.
- Gateway:
  - Manage configurations.
  - Manage instances.
- Instance status transitions are deferred to a later version.

## Provider configuration

```hcl
provider "vidos" {
  region = "eu"        # optional, default eu (service region)

  # required (or set VIDOS_API_KEY)
  api_key = var.vidos_api_key
}
```

Environment variables:

- `VIDOS_API_KEY` (required if `api_key` not set)
- `VIDOS_REGION` (optional)
- `VIDOS_API_VERSION` (optional, default `1`)

## Resources

- `vidos_iam_api_key`
- `vidos_iam_policy`
- `vidos_iam_api_key_policy_attachment`
- `vidos_iam_service_role`
- `vidos_iam_service_role_policy_attachment`
- `vidos_resolver_configuration`
- `vidos_resolver_instance`
- `vidos_verifier_configuration`
- `vidos_verifier_instance`
- `vidos_validator_configuration`
- `vidos_validator_instance`
- `vidos_authorizer_configuration`
- `vidos_authorizer_instance`
- `vidos_gateway_configuration`
- `vidos_gateway_instance`

## Notes

- `vidos_iam_api_key.api_secret` is **write-only**. If an API key is imported, the secret cannot be recovered.
- Attachments fail fast: before attaching, the provider verifies that the policy exists.
- For resources that accept `resource_id`, it is optional and immutable. If omitted, the provider will generate a stable `tf-<hex>` id on create.

## Development

### Local build

```bash
go mod tidy
go build -ldflags "-X main.version=1.0.0" -o ./bin/terraform-provider-vidos .
```

### Unit tests

All unit tests run offline by mocking HTTP using `http.RoundTripper`.

```bash
go test ./...
go test ./... -cover

# Coverage profile + function breakdown
make coverprofile

# Coverage profile + gate (minimum configured in `.testcoverage.yml`)
make coverage
```

## Contributing

Contributions are welcome! Please see the [GitHub repository](https://github.com/vidos-id/terraform-provider-vidos) for more information.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
