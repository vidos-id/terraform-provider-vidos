package main

import "testing"

func TestAPIClient_BaseURLs(t *testing.T) {
	c := &APIClient{cfg: providerConfig{domain: "example.com", defaultRegion: "eu"}}
	if got := c.iamBaseURL(); got != "https://iam.management.global.example.com" {
		t.Fatalf("unexpected iam url: %q", got)
	}
	if got := c.resolverBaseURL(); got != "https://resolver.management.eu.example.com" {
		t.Fatalf("unexpected resolver url: %q", got)
	}
	if got := c.verifierBaseURL(); got != "https://verifier.management.eu.example.com" {
		t.Fatalf("unexpected verifier url: %q", got)
	}
	if got := c.validatorBaseURL(); got != "https://validator.management.eu.example.com" {
		t.Fatalf("unexpected validator url: %q", got)
	}
	if got := c.authorizerBaseURL(); got != "https://authorizer.management.eu.example.com" {
		t.Fatalf("unexpected authorizer url: %q", got)
	}
	if got := c.gatewayBaseURL(); got != "https://gateway.management.eu.example.com" {
		t.Fatalf("unexpected gateway url: %q", got)
	}
}
