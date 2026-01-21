package main

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveOrGenerateResourceID_UnknownFails(t *testing.T) {
	var diags diag.Diagnostics
	_, ok := resolveOrGenerateResourceID(&diags, types.StringUnknown(), path.Root("resource_id"))
	if ok || !diags.HasError() {
		t.Fatalf("expected error")
	}
}

func TestResolveOrGenerateResourceID_NullGenerates(t *testing.T) {
	var diags diag.Diagnostics
	id, ok := resolveOrGenerateResourceID(&diags, types.StringNull(), path.Root("resource_id"))
	if !ok || diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if len(id) < 4 || id[:3] != "tf-" {
		t.Fatalf("unexpected id: %q", id)
	}
}

func TestResolveOrGenerateResourceID_UsesConfigValue(t *testing.T) {
	var diags diag.Diagnostics
	id, ok := resolveOrGenerateResourceID(&diags, types.StringValue("custom"), path.Root("resource_id"))
	if !ok || diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if id != "custom" {
		t.Fatalf("unexpected id: %q", id)
	}
}

func TestGenerateTerraformResourceID_RandReadError(t *testing.T) {
	old := cryptoRandRead
	cryptoRandRead = func([]byte) (int, error) { return 0, errors.New("boom") }
	t.Cleanup(func() { cryptoRandRead = old })

	_, err := generateTerraformResourceID("tf-")
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestResolveOrGenerateResourceID_GenerateErrorAddsDiag(t *testing.T) {
	old := cryptoRandRead
	cryptoRandRead = func([]byte) (int, error) { return 0, errors.New("boom") }
	t.Cleanup(func() { cryptoRandRead = old })

	var diags diag.Diagnostics
	_, ok := resolveOrGenerateResourceID(&diags, types.StringNull(), path.Root("resource_id"))
	if ok {
		t.Fatalf("expected ok=false")
	}
	if !diags.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}
