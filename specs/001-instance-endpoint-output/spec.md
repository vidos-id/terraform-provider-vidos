# Feature Specification: Instance Endpoint Output

**Feature Branch**: `001-instance-endpoint-output`  
**Created**: 2026-01-27  
**Status**: Draft  
**Input**: User description: "add to instance resources a terraform output of endpoint."

## Clarifications

### Session 2026-01-27

- Q: Which format should the resource’s `endpoint` expose? → A: Pass-through: exact string from the platform (no normalization).
- Q: When the platform does NOT provide an endpoint, how should `endpoint` be represented in Terraform state? → A: `null`/unknown (attribute unset until known).
- Q: Which resources should gain the `endpoint` computed attribute? → A: All instance-like resources (any resource representing an instance).
- Q: Should `endpoint` be read-only or configurable? → A: Read-only: `Computed` only.
- Q: Should `endpoint` be marked `Sensitive` in the Terraform schema? → A: No: not `Sensitive` (default).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Export Instance Endpoint (Priority: P1)

As a Terraform user managing Vidos instances, I want the instance resource to expose its reachable endpoint so I can output it for operators and downstream automation.

**Why this priority**: This is the core value: enabling immediate visibility and reuse of the endpoint without manual lookups.

**Independent Test**: Provision a single instance and confirm its endpoint can be surfaced via a Terraform output value.

**Acceptance Scenarios**:

1. **Given** a configuration that provisions an instance, **When** the user applies it, **Then** the instance exposes a non-empty endpoint value when the platform provides one.
2. **Given** a configuration that outputs the instance endpoint, **When** the user applies it, **Then** the output resolves to the same endpoint value exposed by the instance.

---

### User Story 2 - Use Endpoint in Compositions (Priority: P2)

As a module author, I want to reference the instance endpoint from other configurations so I can integrate the instance into dependent workflows.

**Why this priority**: Reuse in compositions is a common next step after exposing the endpoint.

**Independent Test**: Reference the endpoint value from another configuration block and confirm it evaluates without requiring manual data entry.

**Acceptance Scenarios**:

1. **Given** an instance exists, **When** another part of the configuration references its endpoint value, **Then** the plan and apply succeed without additional user-supplied endpoint inputs.

---

### User Story 3 - Endpoint Refresh Behavior (Priority: P3)

As an operator, I want the exposed endpoint value to stay accurate over time so outputs and dependent workflows do not drift from the actual reachable endpoint.

**Why this priority**: Endpoint changes are less frequent than initial provisioning, but correctness over time prevents confusing, stale outputs.

**Independent Test**: Trigger an instance refresh and confirm the exposed endpoint reflects the current value from the platform when it changes.

**Acceptance Scenarios**:

1. **Given** the platform-reported endpoint changes for an existing instance, **When** the user refreshes state and applies any resulting updates, **Then** the exposed endpoint value updates accordingly.

---

### Edge Cases

- Endpoint is not available yet (or is temporarily unknown) after create/read; the resource still completes and the endpoint is represented as `null`/unknown rather than causing a hard failure.
- Endpoint is present but not reachable from the user's network; the system reports the endpoint value as provided without attempting to validate external reachability.
- Endpoint contains no secrets (no embedded credentials/tokens) so it is safe to expose via outputs by default.

## Requirements *(mandatory)*

### Functional Requirements

Assumptions:
- “Endpoint” refers to the platform-reported endpoint value for the instance, exposed as-is (no normalization).
- “Instance resources” refers to the provider resource(s) representing an instance.

- **FR-001**: All instance-like resources (any resource representing an instance) MUST expose an `endpoint` value for each managed instance.
- **FR-002**: The exposed `endpoint` MUST be the platform-reported endpoint value, unmodified, and suitable for use in Terraform outputs and module compositions.
- **FR-003**: If the platform does not provide an endpoint value, the system MUST represent it as `null`/unknown and MUST NOT fail provisioning solely due to its absence.
- **FR-004**: The exposed `endpoint` MUST NOT include credentials, tokens, or other secrets.
- **FR-005**: Documentation MUST describe the endpoint value and include a minimal example showing how users can output it.
- **FR-006**: The `endpoint` attribute MUST be read-only (Computed-only) and MUST NOT be user-configurable.
- **FR-007**: The `endpoint` attribute MUST NOT be marked sensitive.

### Key Entities *(include if feature involves data)*

- **Instance**: A managed unit representing a Vidos instance created/controlled by Terraform.
- **Endpoint**: The externally usable network address associated with an instance.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For instances where the platform provides an endpoint, 100% of applies expose a non-empty endpoint value in state.
- **SC-002**: Users can output the instance endpoint in a single apply without performing any manual lookups.
- **SC-003**: When the platform-reported endpoint changes, users see the updated endpoint after a refresh/apply cycle.
