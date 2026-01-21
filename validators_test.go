package main

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRegionValidator_NullAndUnknown(t *testing.T) {
	v := regionValidator{}
	ctx := context.Background()

	{
		resp := &validator.StringResponse{}
		v.ValidateString(ctx, validator.StringRequest{Path: path.Root("region"), ConfigValue: types.StringNull()}, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected diagnostics for null: %#v", resp.Diagnostics)
		}
	}

	{
		resp := &validator.StringResponse{}
		v.ValidateString(ctx, validator.StringRequest{Path: path.Root("region"), ConfigValue: types.StringUnknown()}, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected diagnostics for unknown: %#v", resp.Diagnostics)
		}
	}
}

func TestRegionValidator_Descriptions(t *testing.T) {
	v := regionValidator{}
	ctx := context.Background()

	d := v.Description(ctx)
	if d == "" {
		t.Fatalf("expected non-empty Description")
	}

	md := v.MarkdownDescription(ctx)
	if md != d {
		t.Fatalf("expected MarkdownDescription to match Description; got %q want %q", md, d)
	}
}

func TestRegionValidator_EmptyInvalid(t *testing.T) {
	v := regionValidator{}
	resp := &validator.StringResponse{}
	v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root("region"), ConfigValue: types.StringValue("")}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}

func TestRegionValidator_InvalidFormats(t *testing.T) {
	cases := []string{"EU", "eu_1", "-eu", "eu-", "eU", "e u"}
	v := regionValidator{}

	for _, tc := range cases {
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root("region"), ConfigValue: types.StringValue(tc)}, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatalf("expected diagnostics error for %q", tc)
		}
	}
}

func TestRegionValidator_ValidSlugs(t *testing.T) {
	cases := []string{"eu", "us", "eu-1", "ap-south-1"}
	v := regionValidator{}

	for _, tc := range cases {
		resp := &validator.StringResponse{}
		v.ValidateString(context.Background(), validator.StringRequest{Path: path.Root("region"), ConfigValue: types.StringValue(tc)}, resp)
		if resp.Diagnostics.HasError() {
			t.Fatalf("unexpected diagnostics for %q: %#v", tc, resp.Diagnostics)
		}
	}
}
