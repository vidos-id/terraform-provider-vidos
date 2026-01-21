package main

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBuildProviderConfig_DefaultsAndApiKeyFromEnv(t *testing.T) {
	oldRegion := os.Getenv("VIDOS_REGION")
	oldKey := os.Getenv("VIDOS_API_KEY")
	t.Cleanup(func() {
		_ = os.Setenv("VIDOS_REGION", oldRegion)
		_ = os.Setenv("VIDOS_API_KEY", oldKey)
	})

	_ = os.Setenv("VIDOS_REGION", "")
	_ = os.Setenv("VIDOS_API_KEY", "  secret  ")

	cfg, diags := buildProviderConfig(providerModel{Region: types.StringNull(), ApiKey: types.StringNull()})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if cfg.domain != defaultDomain {
		t.Fatalf("unexpected domain: %q", cfg.domain)
	}
	if cfg.defaultRegion != "eu" {
		t.Fatalf("expected default region eu, got %q", cfg.defaultRegion)
	}
	if cfg.apiKeySecret != "secret" {
		t.Fatalf("expected api key secret 'secret', got %q", cfg.apiKeySecret)
	}
}

func TestBuildProviderConfig_RegionFromConfigPreferredOverEnv(t *testing.T) {
	oldRegion := os.Getenv("VIDOS_REGION")
	oldKey := os.Getenv("VIDOS_API_KEY")
	t.Cleanup(func() {
		_ = os.Setenv("VIDOS_REGION", oldRegion)
		_ = os.Setenv("VIDOS_API_KEY", oldKey)
	})

	_ = os.Setenv("VIDOS_REGION", "us")
	_ = os.Setenv("VIDOS_API_KEY", "secret")

	cfg, diags := buildProviderConfig(providerModel{Region: types.StringValue(" ap-south-1 "), ApiKey: types.StringNull()})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if cfg.defaultRegion != "ap-south-1" {
		t.Fatalf("expected region ap-south-1, got %q", cfg.defaultRegion)
	}
}

func TestBuildProviderConfig_MissingAPIKey(t *testing.T) {
	oldKey := os.Getenv("VIDOS_API_KEY")
	t.Cleanup(func() { _ = os.Setenv("VIDOS_API_KEY", oldKey) })
	_ = os.Setenv("VIDOS_API_KEY", "")

	_, diags := buildProviderConfig(providerModel{Region: types.StringNull(), ApiKey: types.StringNull()})
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestGetFirstNonEmpty_TrimsAndPrefersConfig(t *testing.T) {
	if got := getFirstNonEmpty(types.StringValue("  hi  "), "env"); got != "hi" {
		t.Fatalf("unexpected value: %q", got)
	}
}

func TestGetFirstNonEmpty_UsesEnvWhenNullOrUnknown(t *testing.T) {
	if got := getFirstNonEmpty(types.StringNull(), "  env  "); got != "env" {
		t.Fatalf("unexpected value: %q", got)
	}
	if got := getFirstNonEmpty(types.StringUnknown(), "env"); got != "env" {
		t.Fatalf("unexpected value: %q", got)
	}
}

func TestBuildManagementBaseURL(t *testing.T) {
	if got := buildManagementBaseURL("iam", "global", "example.com"); got != "https://iam.management.global.example.com" {
		t.Fatalf("unexpected url: %q", got)
	}
}

func TestRequireKnownString(t *testing.T) {
	{
		var diags diag.Diagnostics
		_, ok := requireKnownString(&diags, types.StringUnknown(), path.Root("x"), "x")
		if ok || !diags.HasError() {
			t.Fatalf("expected error for unknown")
		}
	}
	{
		var diags diag.Diagnostics
		_, ok := requireKnownString(&diags, types.StringNull(), path.Root("x"), "x")
		if ok || !diags.HasError() {
			t.Fatalf("expected error for null")
		}
	}
	{
		var diags diag.Diagnostics
		_, ok := requireKnownString(&diags, types.StringValue("   "), path.Root("x"), "x")
		if ok || !diags.HasError() {
			t.Fatalf("expected error for empty")
		}
	}
	{
		var diags diag.Diagnostics
		v, ok := requireKnownString(&diags, types.StringValue("  ok  "), path.Root("x"), "x")
		if !ok || diags.HasError() {
			t.Fatalf("unexpected diagnostics: %#v", diags)
		}
		if v != "ok" {
			t.Fatalf("unexpected value: %q", v)
		}
	}
}

func TestProviderConfigure_SetsClientData(t *testing.T) {
	ctx := context.Background()

	p := &VidosProvider{version: "test"}

	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)

	regionTF, err := types.StringValue("eu").ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("ToTerraformValue error: %s", err)
	}
	keyTF, err := types.StringValue("secret").ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("ToTerraformValue error: %s", err)
	}

	cfg := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			schemaResp.Schema.Type().TerraformType(ctx),
			map[string]tftypes.Value{
				"region":  regionTF,
				"api_key": keyTF,
			},
		),
	}

	var resp provider.ConfigureResponse
	p.Configure(ctx, provider.ConfigureRequest{Config: cfg}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if resp.ResourceData == nil || resp.DataSourceData == nil {
		t.Fatalf("expected provider data to be set")
	}
	if resp.ResourceData != resp.DataSourceData {
		t.Fatalf("expected resource and datasource data to match")
	}
	if _, ok := resp.ResourceData.(*APIClient); !ok {
		t.Fatalf("expected *APIClient, got %T", resp.ResourceData)
	}
}

func TestProviderConfigure_MissingApiKeyAddsError(t *testing.T) {
	ctx := context.Background()

	oldKey := os.Getenv("VIDOS_API_KEY")
	t.Cleanup(func() { _ = os.Setenv("VIDOS_API_KEY", oldKey) })
	_ = os.Setenv("VIDOS_API_KEY", "")

	p := &VidosProvider{version: "test"}

	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)

	regionTF, err := types.StringValue("eu").ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("ToTerraformValue error: %s", err)
	}

	// api_key omitted (null) should be rejected unless env var is set.
	cfg := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(
			schemaResp.Schema.Type().TerraformType(ctx),
			map[string]tftypes.Value{
				"region":  regionTF,
				"api_key": tftypes.NewValue(tftypes.String, nil),
			},
		),
	}

	var resp provider.ConfigureResponse
	p.Configure(ctx, provider.ConfigureRequest{Config: cfg}, &resp)

	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected error diagnostics")
	}
	if resp.ResourceData != nil || resp.DataSourceData != nil {
		t.Fatalf("expected provider data unset on error")
	}
}
