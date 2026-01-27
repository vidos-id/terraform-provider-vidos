# Data Model: Instance Endpoint Output

## Terraform Schema

All instance-like resources MUST expose:

- Attribute: `endpoint`
  - Type: string
  - Computed: true
  - Optional: false
  - Required: false
  - Sensitive: false
  - Description: platform-reported endpoint (pass-through)

## State Model

Extend the shared instance state model to include:

- `Endpoint types.String` (`tfsdk:"endpoint"`)

## API Mapping

### Read

Read responses include an instance object with the JSON field:

- `endpoint` (string)

Mapping rules:

- If `endpoint` exists and is non-empty: set `Endpoint = types.StringValue(endpoint)`.
- If `endpoint` is missing or empty: set `Endpoint = types.StringNull()`.

### Create

Create responses include `instance.endpoint`.

Mapping rules:

- If returned and non-empty: set `Endpoint = types.StringValue(endpoint)`.
- Otherwise: set `Endpoint = types.StringNull()` and rely on the subsequent Read.

### Update

Endpoint is read-only from Terraform; it is refreshed from the platform.

## Notes

- No validation of reachability.
- No normalization of scheme/host/path; provider stores the exact string returned by the API.
