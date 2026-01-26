# Terraform Provider Vidos Constitution

## OpenSpec Instructions

For spec-driven development guidance, see the root [AGENTS.md](../../AGENTS.md) file which contains:
- When to create change proposals vs fix directly
- Proposal structure and workflow
- Spec file format and delta operations
- Validation and archiving processes

## Core Principles

### I. Terraform Provider Best Practices
- Follow HashiCorp's Terraform Plugin Framework patterns
- Resources must be idempotent and support full CRUD lifecycle
- Use schema validators for input validation
- Implement proper import functionality for all resources

### II. Test Coverage
- Unit tests for all resource logic (`*_test.go` files)
- Use test seams (`test_seams.go`) for mocking HTTP interactions
- Acceptance tests should be runnable against real API (when available)
- Table-driven tests preferred for comprehensive coverage

### III. API Client Design
- Centralized client with retry logic and error handling
- Consistent error wrapping with `APIError` type
- Request/response truncation for logging safety
- Configurable timeouts and retry behavior

### IV. Resource Organization
- Configuration resources (`resource_*_configuration.go`) for settings
- Instance resources (`resource_*_instance.go`) for deployable entities
- Common patterns extracted to `resource_*_common.go` files
- IAM resources follow AWS-style patterns (policies, attachments)

### V. Documentation
- Auto-generated docs in `docs/` from schema descriptions
- Examples in `examples/` for each resource type
- README with quickstart and authentication setup

## Development Workflow

1. **Spec First**: For new features, create OpenSpec proposal (see AGENTS.md)
2. **Test First**: Write failing tests before implementation
3. **Validate**: Run `make test` before committing
4. **Document**: Update examples and descriptions with changes

## Governance

- This constitution guides development decisions
- OpenSpec proposals required for new resources or breaking changes
- See [AGENTS.md](../../AGENTS.md) for detailed spec workflow

**Version**: 1.0.0 | **Ratified**: 2026-01-26
