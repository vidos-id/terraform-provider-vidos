package main

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProvider_ResourcesAndDataSources(t *testing.T) {
	p := New().(*VidosProvider)
	if got := len(p.Resources(context.Background())); got == 0 {
		t.Fatalf("expected resources")
	}
	if p.DataSources(context.Background()) != nil {
		t.Fatalf("expected nil data sources")
	}
}

func TestProvider_MetadataAndSchema(t *testing.T) {
	p := &VidosProvider{version: "1.2.3"}

	var meta provider.MetadataResponse
	p.Metadata(context.Background(), provider.MetadataRequest{}, &meta)
	if meta.TypeName != "vidos" {
		t.Fatalf("unexpected type name: %q", meta.TypeName)
	}
	if meta.Version != "1.2.3" {
		t.Fatalf("unexpected version: %q", meta.Version)
	}

	var sch provider.SchemaResponse
	p.Schema(context.Background(), provider.SchemaRequest{}, &sch)
	if sch.Schema.Attributes == nil {
		t.Fatalf("expected schema attributes")
	}
	if _, ok := sch.Schema.Attributes["region"]; !ok {
		t.Fatalf("expected region attribute")
	}
	if _, ok := sch.Schema.Attributes["api_key"]; !ok {
		t.Fatalf("expected api_key attribute")
	}
}

func TestResourceWrappers_MetadataAndSchemaSmoke(t *testing.T) {
	ctx := context.Background()
	providerType := "vidos"

	tests := []struct {
		name string
		new  func() resource.Resource
		suf  string
	}{
		{"iam_api_key", NewIamApiKeyResource, "_iam_api_key"},
		{"iam_policy", NewIamPolicyResource, "_iam_policy"},
		{"iam_api_key_policy_attachment", NewIamApiKeyPolicyAttachmentResource, "_iam_api_key_policy_attachment"},
		{"iam_service_role", NewIamServiceRoleResource, "_iam_service_role"},
		{"iam_service_role_policy_attachment", NewIamServiceRolePolicyAttachmentResource, "_iam_service_role_policy_attachment"},
		{"resolver_configuration", NewResolverConfigurationResource, "_resolver_configuration"},
		{"resolver_instance", NewResolverInstanceResource, "_resolver_instance"},
		{"verifier_configuration", NewVerifierConfigurationResource, "_verifier_configuration"},
		{"verifier_instance", NewVerifierInstanceResource, "_verifier_instance"},
		{"validator_configuration", NewValidatorConfigurationResource, "_validator_configuration"},
		{"validator_instance", NewValidatorInstanceResource, "_validator_instance"},
		{"authorizer_configuration", NewAuthorizerConfigurationResource, "_authorizer_configuration"},
		{"authorizer_instance", NewAuthorizerInstanceResource, "_authorizer_instance"},
		{"gateway_configuration", NewGatewayConfigurationResource, "_gateway_configuration"},
		{"gateway_instance", NewGatewayInstanceResource, "_gateway_instance"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := tc.new()

			var meta resource.MetadataResponse
			r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: providerType}, &meta)
			if meta.TypeName == "" {
				t.Fatalf("expected type name")
			}
			if len(meta.TypeName) < len(providerType)+len(tc.suf) || meta.TypeName[len(meta.TypeName)-len(tc.suf):] != tc.suf {
				t.Fatalf("unexpected type name: %q", meta.TypeName)
			}

			var sch resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &sch)
			if sch.Schema.Attributes == nil {
				t.Fatalf("expected schema")
			}
		})
	}
}
