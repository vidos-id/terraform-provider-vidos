# Feature Specification: Hide Validator From Gateway

**Feature Branch**: `001-hide-gateway-validator`  
**Created**: 2026-01-27  
**Status**: Draft  
**Input**: User description: "gateway-authorizer-validator should not expose the validator in the gateway."

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Validator Not Publicly Reachable (Priority: P1)

As a user of the “Gateway + Authorizer + Validator” reference setup, I want the gateway to avoid exposing a direct route to the validator so that validation is only reachable through the authorizer’s controls.

**Why this priority**: Prevents accidental deployment of a public “bypass” path that undermines expected authorization behavior.

**Independent Test**: Review the gateway’s configured routes and confirm there is no route that forwards requests directly to the validator.

**Acceptance Scenarios**:

1. **Given** the reference setup is configured, **When** the gateway routing rules are inspected, **Then** no route exists that targets the validator service directly.
2. **Given** the reference setup is configured, **When** a request is made to the previously documented direct validation route, **Then** the gateway does not forward the request to the validator and the request fails with a “not found” or “not allowed” outcome.
3. **Given** the reference setup is configured, **When** a validation request is made via the authorizer route, **Then** the request can be handled successfully according to the authorizer’s policies.

---

### User Story 2 - Validator Still Usable Behind Authorizer (Priority: P2)

As a user of the reference setup, I want the authorizer to continue using the validator for validation behind the scenes, without the gateway needing direct access to the validator.

**Why this priority**: Keeps the “authorizer validates via validator” capability intact while reducing public exposure.

**Independent Test**: Configure only the authorizer-to-validator relationship and confirm validation continues to function via the authorizer route.

**Acceptance Scenarios**:

1. **Given** the gateway has no direct validator route, **When** a validation request is processed through the authorizer route, **Then** the authorizer can still validate using the validator and return the expected outcome.
2. **Given** the reference setup is configured, **When** permissions/roles are reviewed, **Then** the gateway does not require permissions intended specifically for gateway-to-validator calls.

---

### User Story 3 - Clear Documentation and Migration Note (Priority: P3)

As a reader of the guide for this reference setup, I want the documentation to describe the new routing model (no direct validator exposure) and clearly explain how to adapt if I previously used the direct validator route.

**Why this priority**: Avoids confusion and reduces support load for users following older instructions.

**Independent Test**: Read the guide end-to-end and verify it does not suggest a direct validator route and includes a short migration note.

**Acceptance Scenarios**:

1. **Given** the guide is published, **When** a reader follows it to configure routing, **Then** the documented gateway paths do not include any direct-to-validator route.
2. **Given** a reader previously relied on a direct validator route, **When** they read the migration note, **Then** they can identify what changed and what alternative path to use.

---

### Edge Cases

- Requests to the old direct validator path still occur (automation, bookmarks): they must fail safely (not forwarded to validator).
- Users intentionally wanted public direct validation: guide must make it clear this is no longer part of the reference setup.
- Partial adoption (only docs updated, not the example configuration, or vice versa) must be avoided; both must align.

## Requirements *(mandatory)*

### Functional Requirements

Assumptions:

- This change applies to the “Gateway + Authorizer + Validator” reference setup (example + guide), not to all possible user deployments.
- The validator is intended to be used by the authorizer as a backing capability, not as a publicly exposed gateway route in this reference setup.

Out of Scope:

- Changing the underlying services’ behavior, security model, or validation logic.
- Removing or changing the validator resource itself from the example.
- Preventing users from creating their own gateway-to-validator routes in their own deployments (this spec only changes the published reference setup).

- **FR-001**: The reference gateway configuration MUST NOT include any publicly reachable route that forwards requests directly to the validator service.
- **FR-002**: The reference setup MUST continue to support validation via the authorizer path, with the authorizer able to use the validator as needed.
- **FR-003**: The guide for the reference setup MUST NOT describe or recommend a direct gateway-to-validator route.
- **FR-004**: The guide MUST include a brief migration note stating that any previously documented direct validator path is removed/unsupported in this reference setup and that validation should be performed via the authorizer route.
- **FR-005**: The reference setup MUST NOT require gateway-specific permissions intended solely for calling the validator directly.

### Non-Goals

- Deprecating or removing validator functionality.
- Introducing new validation endpoints or new user-facing features beyond the reference setup.

### Key Entities *(include if feature involves data)*

- **Gateway Route**: A publicly reachable path on the gateway that forwards requests to a backing service.
- **Authorizer Service**: The service responsible for applying authorization/validation policy before returning a result.
- **Validator Service**: The backing service used to perform validation checks on behalf of the authorizer.
- **Reference Setup (Example + Guide)**: The published configuration and documentation users follow to deploy the pattern.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The published reference gateway routing includes 0 direct-to-validator routes.
- **SC-002**: A basic verification demonstrates that at least 5 representative validation requests succeed via the authorizer route while equivalent requests to the old direct validation route fail safely (not forwarded).
- **SC-003**: Documentation for the reference setup contains 0 mentions of a direct gateway-to-validator path and includes a migration note.

## Dependencies

- Existing reference example and guide content for “Gateway + Authorizer + Validator”.
