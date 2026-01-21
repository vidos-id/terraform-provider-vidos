package main

import "testing"

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
