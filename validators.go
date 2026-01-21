package main

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = (*regionValidator)(nil)

var regionSlugRegex = regexp.MustCompile(`^[a-z](?:[a-z0-9-]*[a-z0-9])?$`)

type regionValidator struct{}

func (v regionValidator) Description(_ context.Context) string {
	return "Region must be a lowercase slug like 'eu' (letters, digits, hyphens)."
}

func (v regionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v regionValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid region", "region must not be empty when set")
		return
	}

	// Accept common slug format.
	// Examples: eu, us, uk
	if !regionSlugRegex.MatchString(value) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid region", "region must be lowercase and contain only letters, digits, and hyphens")
		return
	}
}
