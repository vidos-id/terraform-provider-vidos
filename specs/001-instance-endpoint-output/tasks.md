---

description: "Tasks for Instance Endpoint Output"
---

# Tasks: Instance Endpoint Output

**Input**: Design documents from `/Users/robdefeo/Documents/GitHub/terraform-provider-vidos/specs/001-instance-endpoint-output/`

**Tech Stack**: Go 1.22, Terraform Plugin Framework (`github.com/hashicorp/terraform-plugin-framework v1.10.0`)

**Scope**:
- Add a computed, non-sensitive `endpoint` attribute to all instance-like resources (gateway/authorizer/validator/verifier/resolver instances).
- Map `endpoint` from the management API JSON field `endpoint` (pass-through; no normalization).
- If `endpoint` is missing/empty in the API response, set Terraform state to `null` and do not fail.

**Tests**: Included (spec marks testing as mandatory; plan requires updating unit tests).

## Format

Every task uses:

`- [ ] T### [P?] [US#?] Description with file path`

- `[P]` means it can be done in parallel with other `[P]` tasks (different files; no dependency on incomplete work).
- `[US#]` labels appear only in user story phases.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Ensure local prerequisites and baseline verification commands are known.

- [X] T001 Confirm Go/tooling expectations in `go.mod` and `Makefile`
- [X] T002 Run `.specify/scripts/bash/check-prerequisites.sh --json` and confirm FEATURE_DIR matches `specs/001-instance-endpoint-output/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared model + mapping primitives required by all instance resources.

**⚠️ CRITICAL**: Do not start user story work until these are done.

- [X] T003 Extend shared instance state model with `Endpoint types.String` in `resource_instance_common.go`
- [X] T004 Extend instance API response structs to include `endpoint` JSON field in `resource_instance_common.go`
- [X] T005 Add a shared mapping helper for endpoint string → `types.StringValue`/`types.StringNull` in `resource_instance_common.go`
- [X] T006 Update shared test helpers/fixtures to include endpoint where relevant in `tfsdk_test_helpers.go`

**Checkpoint**: Foundation ready - user story implementation can begin.

---

## Phase 3: User Story 1 - Export Instance Endpoint (Priority: P1) MVP

**Goal**: Instance resources expose a computed `endpoint` attribute suitable for outputs.

**Independent Test**: Provision a single instance and confirm its endpoint can be surfaced via a Terraform output value.

### Implementation (and required unit tests)

- [X] T007 [US1] Add `endpoint` as Computed-only, non-sensitive schema attribute in `resource_instance_common.go`
- [X] T008 [P] [US1] Expose `endpoint` in gateway instance schema in `resource_gateway_instance.go`
- [X] T009 [P] [US1] Expose `endpoint` in authorizer instance schema in `resource_authorizer_instance.go`
- [X] T010 [P] [US1] Expose `endpoint` in validator instance schema in `resource_validator_instance.go`
- [X] T011 [P] [US1] Expose `endpoint` in verifier instance schema in `resource_verifier_instance.go`
- [X] T012 [P] [US1] Expose `endpoint` in resolver instance schema in `resource_resolver_instance.go`
- [X] T013 [US1] Map API `endpoint` into Terraform state during Create/Read in `resource_instance_resource.go`
- [X] T014 [US1] Add unit tests covering endpoint present mapping in `resource_instance_resource_test.go`
- [X] T015 [US1] Add schema tests asserting `endpoint` is Computed and not Sensitive in `resource_instance_wrappers_test.go`
- [X] T016 [US1] Run `make test` (see `Makefile`) and ensure US1 scenarios pass
---

## Phase 4: User Story 2 - Use Endpoint in Compositions (Priority: P2)

**Goal**: Users can reference `endpoint` in outputs/modules and learn how via docs/examples.

**Independent Test**: Reference the endpoint value from another configuration block and confirm it evaluates without requiring manual data entry.

### Documentation and examples

- [X] T017 [P] [US2] Document `endpoint` attribute and output usage example in `docs/resources/gateway_instance.md`
- [X] T018 [P] [US2] Document `endpoint` attribute and output usage example in `docs/resources/authorizer_instance.md`
- [X] T019 [P] [US2] Document `endpoint` attribute and output usage example in `docs/resources/validator_instance.md`
- [X] T020 [P] [US2] Document `endpoint` attribute and output usage example in `docs/resources/verifier_instance.md`
- [X] T021 [P] [US2] Document `endpoint` attribute and output usage example in `docs/resources/resolver_instance.md`
- [X] T022 [US2] Add a minimal endpoint output example in `examples/gateway-authorizer/outputs.tf` and `examples/gateway-authorizer-validator/outputs.tf`
- [X] T023 [US2] Align feature quickstart snippet with docs/examples in `specs/001-instance-endpoint-output/quickstart.md`

---

## Phase 5: User Story 3 - Endpoint Refresh Behavior (Priority: P3)

**Goal**: Endpoint in state stays accurate after refresh/apply cycles.

**Independent Test**: Trigger an instance refresh and confirm the exposed endpoint reflects the current value from the platform when it changes.

### Unit tests for refresh semantics

- [X] T024 [US3] Add unit test: missing/empty API endpoint maps to `null` on refresh in `resource_instance_resource_test.go`
- [X] T025 [US3] Add unit test: endpoint value updates when API returns a new endpoint on refresh in `resource_instance_resource_test.go`
- [X] T026 [US3] Run `make test` (see `Makefile`) and ensure US3 scenarios pass

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Consistency, formatting, and final verification.

- [X] T027 [P] Run `gofmt` on `resource_instance_common.go`, `resource_instance_resource.go`, `resource_*_instance.go`, `resource_instance_*_test.go`
- [X] T028 [P] Run `make test` and optionally `make coverprofile` (see `Makefile`)
- [X] T029 [P] Verify docs/examples match FR-001..FR-007 in `specs/001-instance-endpoint-output/spec.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- Phase 1 (Setup) → Phase 2 (Foundational) → Phase 3 (US1) → Phase 4 (US2) and Phase 5 (US3) → Phase 6 (Polish)

### User Story Dependencies

- US1 (P1) is the prerequisite for US2 and US3 (they depend on the attribute existing and being mapped).

---

## Parallel Execution Examples

### User Story 1

```bash
# Can be done in parallel after T007:
Task: "T008 Expose endpoint in resource_gateway_instance.go"
Task: "T009 Expose endpoint in resource_authorizer_instance.go"
Task: "T010 Expose endpoint in resource_validator_instance.go"
Task: "T011 Expose endpoint in resource_verifier_instance.go"
Task: "T012 Expose endpoint in resource_resolver_instance.go"
```

### User Story 2

```bash
# Documentation updates can be done in parallel:
Task: "T017 Update docs/resources/gateway_instance.md"
Task: "T018 Update docs/resources/authorizer_instance.md"
Task: "T019 Update docs/resources/validator_instance.md"
Task: "T020 Update docs/resources/verifier_instance.md"
Task: "T021 Update docs/resources/resolver_instance.md"
```

### User Story 3

```bash
# No meaningful parallelism inside US3 (both tests touch the same file):
Task: "T024 Add null-on-missing endpoint refresh test"
Task: "T025 Add endpoint-update refresh test"
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1 + Phase 2
2. Implement Phase 3 (US1)
3. Run `make test` and validate the US1 independent test using the example in `specs/001-instance-endpoint-output/quickstart.md`
4. Stop and verify before starting docs polish (US2) and refresh semantics (US3)

### Incremental Delivery

1. US1: attribute exists + mapped + unit tests
2. US2: docs/examples teach composition usage
3. US3: refresh behavior covered by unit tests
4. Polish: formatting + final validation
