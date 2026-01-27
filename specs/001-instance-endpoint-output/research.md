# Research: Instance Endpoint Output

## Goal

Identify the API field for instance endpoint and define Terraform mapping behavior for a computed `endpoint` attribute.

## Findings

### OpenAPI field mapping

Source: user-provided OpenAPI 3.0.2 snippet for the Vidos verifier management API.

- `GET /instances` returns an array of instance objects at `instances[*]`.
- Each instance object includes a top-level JSON field `endpoint`:
  - `endpoint: string` with `format: uri`
  - listed under `required` for the instance schema in the OpenAPI snippet.
- `POST /instances` (create) returns `{ instance: { ... endpoint } }`.
- `GET /instances/{instanceResourceId}` (read) returns `{ instance: { ... endpoint } }`.

Decision: Treat the platform endpoint as `endpoint` (exact field name) and map it as-is.

### Terraform representation for “unset”

Spec requirement: if the platform does not provide an endpoint, represent it as `null`/unknown and do not fail.

Decision: Represent an absent/empty endpoint as `types.StringNull()`.

Rationale:
- Matches “attribute unset” semantics in Terraform state.
- Avoids propagating "unknown" indefinitely; provider can always set null when the API does not return a value.

Alternatives considered:
- `types.StringUnknown()` when endpoint missing: rejected because missing/absent from API is not inherently "unknown" to Terraform; it is "not provided" and should stabilize to null.

### When is endpoint available?

The OpenAPI snippet includes `endpoint` in create and read responses.

Decision:
- Populate `endpoint` during Create if present in the create response; otherwise it will be populated on the subsequent Read.

Rationale:
- Improves UX (endpoint appears in state immediately when the API returns it).
- Still correct if create response omits it in practice; read will converge state.
