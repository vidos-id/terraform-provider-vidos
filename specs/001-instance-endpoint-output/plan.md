# Implementation Plan: Instance Endpoint Output

**Branch**: `001-instance-endpoint-output` | **Date**: 2026-01-27 | **Spec**: `/Users/robdefeo/Documents/GitHub/terraform-provider-vidos/specs/001-instance-endpoint-output/spec.md`
**Input**: Feature specification from `/Users/robdefeo/Documents/GitHub/terraform-provider-vidos/specs/001-instance-endpoint-output/spec.md`

## Summary

- Add a computed, non-sensitive `endpoint` attribute to all instance-like resources (gateway/authorizer/validator/verifier/resolver instances).
- Map `endpoint` directly from the management API JSON field `endpoint` (pass-through; no normalization).
- If `endpoint` is missing/empty in the API response, set Terraform state to `null` for `endpoint` and do not fail.

## Technical Context

**Language/Version**: Go 1.22
**Primary Dependencies**: Terraform Plugin Framework (`github.com/hashicorp/terraform-plugin-framework v1.10.0`)
**Storage**: N/A (Terraform provider; state managed by Terraform)
**Testing**: `go test` (via `make test`)
**Target Platform**: darwin/linux/windows (Terraform provider binary)
**Project Type**: Single Go module Terraform provider
**Performance Goals**: N/A (single attribute mapping; negligible overhead)
**Constraints**: Maintain idempotent CRUD; do not error when `endpoint` absent
**Scale/Scope**: Add one attribute across instance-like resources + tests + docs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- Use Terraform Plugin Framework patterns; `endpoint` is Computed-only and not Sensitive.
- CRUD remains idempotent; import continues to work for instance resources.
- Update unit tests to cover endpoint present and endpoint missing/empty => null.
- Update docs for all instance-like resources and add a minimal example output.

## Project Structure

### Documentation (this feature)

```text
specs/001-instance-endpoint-output/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
.
├── resource_instance_common.go
├── resource_instance_resource.go
├── resource_instance_resource_test.go
├── resource_gateway_instance.go
├── resource_authorizer_instance.go
├── resource_validator_instance.go
├── resource_verifier_instance.go
├── resource_resolver_instance.go
├── tfsdk_test_helpers.go
├── docs/
│   └── resources/
│       ├── gateway_instance.md
│       ├── authorizer_instance.md
│       ├── validator_instance.md
│       ├── verifier_instance.md
│       └── resolver_instance.md
└── examples/
    └── vidos-demo-account/
        └── outputs.tf
```

**Structure Decision**: Single Go module Terraform provider; instance resource wrappers define schema and delegate CRUD to shared instance implementation.

## Phase 0: Research

### API Field Mapping

- Source of truth: provided OpenAPI 3.0.2 snippet.
- Instance object includes top-level JSON field `endpoint` (string, `format: uri`).
- Create: `POST /instances` response includes `instance.endpoint`.
- Read: `GET /instances/{instanceResourceId}` response includes `instance.endpoint`.
- List: `GET /instances` response includes `instances[*].endpoint`.

### Terraform State Representation

- Schema: `endpoint` is `Computed: true`, `Optional: false`, `Sensitive: false`.
- Model: `Endpoint types.String`.
- Mapping:
  - If API value is non-empty: `types.StringValue(endpoint)`.
  - If API value is missing or empty: `types.StringNull()`.

## Phase 1: Design

- Add `endpoint` to shared `instanceModel` and to the JSON response struct(s) used for reads.
- Extend all instance-like resource schemas to include `endpoint`.
- Ensure refresh behavior: read path always maps latest API value into state.

## Phase 2: Implementation Tasks (preview; created later in tasks.md)

- Update schema for all instance-like resources to include `endpoint`.
- Extend shared model + response unmarshalling + state mapping.
- Update unit tests and test schema helpers.
- Update docs and example outputs.

## Verification

- Run: `make test`
- Optional: `make coverprofile`
