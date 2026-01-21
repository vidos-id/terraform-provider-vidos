package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func TestJoinURL(t *testing.T) {
	if got := joinURL("https://example.com/", "/x"); got != "https://example.com/x" {
		t.Fatalf("unexpected url: %q", got)
	}
}

func TestJoinURLWithQuery_InvalidBaseFallsBack(t *testing.T) {
	if got := joinURLWithQuery(":// bad", "/x", map[string]string{"a": "b"}); got != ":// bad/x" {
		t.Fatalf("unexpected url: %q", got)
	}
}

func TestJoinURLWithQuery_SkipsEmptyValuesAndEncodes(t *testing.T) {
	got := joinURLWithQuery("https://example.com", "/x", map[string]string{"a": "b", "empty": "  "})
	if got != "https://example.com/x?a=b" {
		t.Fatalf("unexpected url: %q", got)
	}
}

func TestParseJSONToAny_EmptyInput(t *testing.T) {
	var diags diag.Diagnostics
	out := parseJSONToAny(&diags, "   ", path.Root("document"), "document")
	if out != nil {
		t.Fatalf("expected nil output")
	}
	if !diags.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}

func TestParseJSONToAny_InvalidJSON(t *testing.T) {
	var diags diag.Diagnostics
	out := parseJSONToAny(&diags, "{", path.Root("document"), "document")
	if out != nil {
		t.Fatalf("expected nil output")
	}
	if !diags.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}

func TestParseJSONToAny_ValidJSON(t *testing.T) {
	var diags diag.Diagnostics
	out := parseJSONToAny(&diags, `{"a":1}`, path.Root("document"), "document")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if out == nil {
		t.Fatalf("expected output")
	}

	m, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("expected map output, got %T", out)
	}
	if m["a"] != float64(1) {
		t.Fatalf("unexpected value: %#v", m["a"])
	}
}
