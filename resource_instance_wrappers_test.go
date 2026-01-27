package main

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestInstanceWrappers_BaseURLWiring(t *testing.T) {
	client := &APIClient{cfg: providerConfig{domain: "example.com", defaultRegion: "eu"}}

	{
		r := NewGatewayInstanceResource().(*GatewayInstanceResource)
		if r.baseURL == nil {
			t.Fatalf("expected gateway baseURL function")
		}
		if got := r.baseURL(client); got != "https://gateway.management.eu.example.com" {
			t.Fatalf("unexpected gateway baseURL: %q", got)
		}
	}

	{
		r := NewAuthorizerInstanceResource().(*AuthorizerInstanceResource)
		if r.baseURL == nil {
			t.Fatalf("expected authorizer baseURL function")
		}
		if got := r.baseURL(client); got != "https://authorizer.management.eu.example.com" {
			t.Fatalf("unexpected authorizer baseURL: %q", got)
		}
	}

	{
		r := NewResolverInstanceResource().(*ResolverInstanceResource)
		if r.baseURL == nil {
			t.Fatalf("expected resolver baseURL function")
		}
		if got := r.baseURL(client); got != "https://resolver.management.eu.example.com" {
			t.Fatalf("unexpected resolver baseURL: %q", got)
		}
	}

	{
		r := NewValidatorInstanceResource().(*ValidatorInstanceResource)
		if r.baseURL == nil {
			t.Fatalf("expected validator baseURL function")
		}
		if got := r.baseURL(client); got != "https://validator.management.eu.example.com" {
			t.Fatalf("unexpected validator baseURL: %q", got)
		}
	}

	{
		r := NewVerifierInstanceResource().(*VerifierInstanceResource)
		if r.baseURL == nil {
			t.Fatalf("expected verifier baseURL function")
		}
		if got := r.baseURL(client); got != "https://verifier.management.eu.example.com" {
			t.Fatalf("unexpected verifier baseURL: %q", got)
		}
	}
}

func TestInstanceWrappers_EndpointSchemaIsComputedNotSensitive(t *testing.T) {
	resources := []struct {
		name string
		r    func() schema.Schema
	}{
		{"gateway", func() schema.Schema {
			var resp resource.SchemaResponse
			NewGatewayInstanceResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
			return resp.Schema
		}},
		{"authorizer", func() schema.Schema {
			var resp resource.SchemaResponse
			NewAuthorizerInstanceResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
			return resp.Schema
		}},
		{"validator", func() schema.Schema {
			var resp resource.SchemaResponse
			NewValidatorInstanceResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
			return resp.Schema
		}},
		{"verifier", func() schema.Schema {
			var resp resource.SchemaResponse
			NewVerifierInstanceResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
			return resp.Schema
		}},
		{"resolver", func() schema.Schema {
			var resp resource.SchemaResponse
			NewResolverInstanceResource().Schema(context.Background(), resource.SchemaRequest{}, &resp)
			return resp.Schema
		}},
	}

	for _, tt := range resources {
		s := tt.r()
		attr, ok := s.Attributes["endpoint"]
		if !ok {
			t.Fatalf("%s: expected endpoint attribute", tt.name)
		}

		stringAttr, ok := attr.(schema.StringAttribute)
		if !ok {
			t.Fatalf("%s: expected endpoint to be StringAttribute", tt.name)
		}
		if !stringAttr.Computed {
			t.Fatalf("%s: expected endpoint Computed", tt.name)
		}
		if stringAttr.Sensitive {
			t.Fatalf("%s: expected endpoint not Sensitive", tt.name)
		}
	}
}
