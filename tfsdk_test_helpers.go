package main

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func instanceSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id":               schema.StringAttribute{Optional: true, Computed: true},
			"name":                      schema.StringAttribute{Required: true},
			"configuration_resource_id": schema.StringAttribute{Optional: true},
			"inline_configuration":      schema.StringAttribute{Optional: true, Computed: true},
			"endpoint":                  schema.StringAttribute{Computed: true},
		},
	}
}

func instanceConfig(t *testing.T, v instanceModel) tfsdk.Config {
	t.Helper()
	ctx := context.Background()

	s := instanceSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id":               tftypes.String,
		"name":                      tftypes.String,
		"configuration_resource_id": tftypes.String,
		"inline_configuration":      tftypes.String,
		"endpoint":                  tftypes.String,
	}

	ridTF, err := v.ResourceID.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("resource_id ToTerraformValue: %s", err)
	}
	nameTF, err := v.Name.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("name ToTerraformValue: %s", err)
	}
	cidTF, err := v.ConfigurationResourceID.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("configuration_resource_id ToTerraformValue: %s", err)
	}
	inlineTF, err := v.InlineConfiguration.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("inline_configuration ToTerraformValue: %s", err)
	}
	endpointTF, err := v.Endpoint.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("endpoint ToTerraformValue: %s", err)
	}

	return tfsdk.Config{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id":               ridTF,
				"name":                      nameTF,
				"configuration_resource_id": cidTF,
				"inline_configuration":      inlineTF,
				"endpoint":                  endpointTF,
			},
		),
	}
}

func instancePlan(t *testing.T, v instanceModel) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()

	s := instanceSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id":               tftypes.String,
		"name":                      tftypes.String,
		"configuration_resource_id": tftypes.String,
		"inline_configuration":      tftypes.String,
		"endpoint":                  tftypes.String,
	}

	ridTF, err := v.ResourceID.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("resource_id ToTerraformValue: %s", err)
	}
	nameTF, err := v.Name.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("name ToTerraformValue: %s", err)
	}
	cidTF, err := v.ConfigurationResourceID.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("configuration_resource_id ToTerraformValue: %s", err)
	}
	inlineTF, err := v.InlineConfiguration.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("inline_configuration ToTerraformValue: %s", err)
	}
	endpointTF, err := v.Endpoint.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("endpoint ToTerraformValue: %s", err)
	}

	return tfsdk.Plan{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id":               ridTF,
				"name":                      nameTF,
				"configuration_resource_id": cidTF,
				"inline_configuration":      inlineTF,
				"endpoint":                  endpointTF,
			},
		),
	}
}

func instanceState(t *testing.T, v instanceModel) tfsdk.State {
	t.Helper()
	ctx := context.Background()

	s := instanceSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id":               tftypes.String,
		"name":                      tftypes.String,
		"configuration_resource_id": tftypes.String,
		"inline_configuration":      tftypes.String,
		"endpoint":                  tftypes.String,
	}

	ridTF, err := v.ResourceID.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("resource_id ToTerraformValue: %s", err)
	}
	nameTF, err := v.Name.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("name ToTerraformValue: %s", err)
	}
	cidTF, err := v.ConfigurationResourceID.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("configuration_resource_id ToTerraformValue: %s", err)
	}
	inlineTF, err := v.InlineConfiguration.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("inline_configuration ToTerraformValue: %s", err)
	}
	endpointTF, err := v.Endpoint.ToTerraformValue(ctx)
	if err != nil {
		t.Fatalf("endpoint ToTerraformValue: %s", err)
	}

	return tfsdk.State{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id":               ridTF,
				"name":                      nameTF,
				"configuration_resource_id": cidTF,
				"inline_configuration":      inlineTF,
				"endpoint":                  endpointTF,
			},
		),
	}
}

func mustTerraformValue(t *testing.T, v types.String) tftypes.Value {
	t.Helper()
	out, err := v.ToTerraformValue(context.Background())
	if err != nil {
		t.Fatalf("ToTerraformValue error: %s", err)
	}
	return out
}

func configurationSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{Optional: true, Computed: true},
			"name":        schema.StringAttribute{Required: true},
			"values":      schema.StringAttribute{Required: true},
		},
	}
}

func configurationConfig(t *testing.T, v configurationModel) tfsdk.Config {
	t.Helper()

	s := configurationSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id": tftypes.String,
		"name":        tftypes.String,
		"values":      tftypes.String,
	}

	return tfsdk.Config{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id": mustTerraformValue(t, v.ResourceID),
				"name":        mustTerraformValue(t, v.Name),
				"values":      mustTerraformValue(t, v.Values),
			},
		),
	}
}

func configurationPlan(t *testing.T, v configurationModel) tfsdk.Plan {
	t.Helper()

	s := configurationSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id": tftypes.String,
		"name":        tftypes.String,
		"values":      tftypes.String,
	}

	return tfsdk.Plan{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id": mustTerraformValue(t, v.ResourceID),
				"name":        mustTerraformValue(t, v.Name),
				"values":      mustTerraformValue(t, v.Values),
			},
		),
	}
}

func configurationState(t *testing.T, v configurationModel) tfsdk.State {
	t.Helper()

	s := configurationSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id": tftypes.String,
		"name":        tftypes.String,
		"values":      tftypes.String,
	}

	return tfsdk.State{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id": mustTerraformValue(t, v.ResourceID),
				"name":        mustTerraformValue(t, v.Name),
				"values":      mustTerraformValue(t, v.Values),
			},
		),
	}
}

func iamApiKeySchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id":            schema.StringAttribute{Computed: true},
			"name":                   schema.StringAttribute{Required: true},
			"inline_policy_document": schema.StringAttribute{Optional: true},
			"api_secret":             schema.StringAttribute{Computed: true, Sensitive: true},
		},
	}
}

func iamApiKeyPlan(t *testing.T, v iamApiKeyModel) tfsdk.Plan {
	t.Helper()

	s := iamApiKeySchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id":            tftypes.String,
		"name":                   tftypes.String,
		"inline_policy_document": tftypes.String,
		"api_secret":             tftypes.String,
	}

	return tfsdk.Plan{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id":            mustTerraformValue(t, v.ResourceID),
				"name":                   mustTerraformValue(t, v.Name),
				"inline_policy_document": mustTerraformValue(t, v.InlinePolicyDocument),
				"api_secret":             mustTerraformValue(t, v.ApiSecret),
			},
		),
	}
}

func iamApiKeyState(t *testing.T, v iamApiKeyModel) tfsdk.State {
	t.Helper()

	s := iamApiKeySchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id":            tftypes.String,
		"name":                   tftypes.String,
		"inline_policy_document": tftypes.String,
		"api_secret":             tftypes.String,
	}

	return tfsdk.State{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id":            mustTerraformValue(t, v.ResourceID),
				"name":                   mustTerraformValue(t, v.Name),
				"inline_policy_document": mustTerraformValue(t, v.InlinePolicyDocument),
				"api_secret":             mustTerraformValue(t, v.ApiSecret),
			},
		),
	}
}

func iamApiKeyPolicyAttachmentSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true},
			"api_key_id":  schema.StringAttribute{Required: true},
			"policy_type": schema.StringAttribute{Required: true},
			"policy_id":   schema.StringAttribute{Required: true},
		},
	}
}

func iamApiKeyPolicyAttachmentPlan(t *testing.T, v iamApiKeyPolicyAttachmentModel) tfsdk.Plan {
	t.Helper()

	s := iamApiKeyPolicyAttachmentSchema()
	attrTypes := map[string]tftypes.Type{
		"id":          tftypes.String,
		"api_key_id":  tftypes.String,
		"policy_type": tftypes.String,
		"policy_id":   tftypes.String,
	}

	return tfsdk.Plan{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"id":          mustTerraformValue(t, v.ID),
				"api_key_id":  mustTerraformValue(t, v.ApiKeyID),
				"policy_type": mustTerraformValue(t, v.PolicyType),
				"policy_id":   mustTerraformValue(t, v.PolicyID),
			},
		),
	}
}

func iamApiKeyPolicyAttachmentState(t *testing.T, v iamApiKeyPolicyAttachmentModel) tfsdk.State {
	t.Helper()

	s := iamApiKeyPolicyAttachmentSchema()
	attrTypes := map[string]tftypes.Type{
		"id":          tftypes.String,
		"api_key_id":  tftypes.String,
		"policy_type": tftypes.String,
		"policy_id":   tftypes.String,
	}

	return tfsdk.State{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"id":          mustTerraformValue(t, v.ID),
				"api_key_id":  mustTerraformValue(t, v.ApiKeyID),
				"policy_type": mustTerraformValue(t, v.PolicyType),
				"policy_id":   mustTerraformValue(t, v.PolicyID),
			},
		),
	}
}

func iamPolicySchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{Optional: true, Computed: true},
			"name":        schema.StringAttribute{Required: true},
			"document":    schema.StringAttribute{Required: true},
		},
	}
}

func iamPolicyConfig(t *testing.T, v iamPolicyModel) tfsdk.Config {
	t.Helper()

	s := iamPolicySchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id": tftypes.String,
		"name":        tftypes.String,
		"document":    tftypes.String,
	}

	return tfsdk.Config{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id": mustTerraformValue(t, v.ResourceID),
				"name":        mustTerraformValue(t, v.Name),
				"document":    mustTerraformValue(t, v.Document),
			},
		),
	}
}

func iamPolicyPlan(t *testing.T, v iamPolicyModel) tfsdk.Plan {
	t.Helper()

	s := iamPolicySchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id": tftypes.String,
		"name":        tftypes.String,
		"document":    tftypes.String,
	}

	return tfsdk.Plan{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id": mustTerraformValue(t, v.ResourceID),
				"name":        mustTerraformValue(t, v.Name),
				"document":    mustTerraformValue(t, v.Document),
			},
		),
	}
}

func iamPolicyState(t *testing.T, v iamPolicyModel) tfsdk.State {
	t.Helper()

	s := iamPolicySchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id": tftypes.String,
		"name":        tftypes.String,
		"document":    tftypes.String,
	}

	return tfsdk.State{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id": mustTerraformValue(t, v.ResourceID),
				"name":        mustTerraformValue(t, v.Name),
				"document":    mustTerraformValue(t, v.Document),
			},
		),
	}
}

func iamServiceRoleSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id":            schema.StringAttribute{Optional: true, Computed: true},
			"name":                   schema.StringAttribute{Required: true},
			"inline_policy_document": schema.StringAttribute{Optional: true},
		},
	}
}

func iamServiceRoleConfig(t *testing.T, v iamServiceRoleModel) tfsdk.Config {
	t.Helper()

	s := iamServiceRoleSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id":            tftypes.String,
		"name":                   tftypes.String,
		"inline_policy_document": tftypes.String,
	}

	return tfsdk.Config{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id":            mustTerraformValue(t, v.ResourceID),
				"name":                   mustTerraformValue(t, v.Name),
				"inline_policy_document": mustTerraformValue(t, v.InlinePolicyDocument),
			},
		),
	}
}

func iamServiceRolePlan(t *testing.T, v iamServiceRoleModel) tfsdk.Plan {
	t.Helper()

	s := iamServiceRoleSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id":            tftypes.String,
		"name":                   tftypes.String,
		"inline_policy_document": tftypes.String,
	}

	return tfsdk.Plan{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id":            mustTerraformValue(t, v.ResourceID),
				"name":                   mustTerraformValue(t, v.Name),
				"inline_policy_document": mustTerraformValue(t, v.InlinePolicyDocument),
			},
		),
	}
}

func iamServiceRoleState(t *testing.T, v iamServiceRoleModel) tfsdk.State {
	t.Helper()

	s := iamServiceRoleSchema()
	attrTypes := map[string]tftypes.Type{
		"resource_id":            tftypes.String,
		"name":                   tftypes.String,
		"inline_policy_document": tftypes.String,
	}

	return tfsdk.State{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"resource_id":            mustTerraformValue(t, v.ResourceID),
				"name":                   mustTerraformValue(t, v.Name),
				"inline_policy_document": mustTerraformValue(t, v.InlinePolicyDocument),
			},
		),
	}
}

func iamServiceRolePolicyAttachmentSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Computed: true},
			"service_role_id": schema.StringAttribute{Required: true},
			"policy_type":     schema.StringAttribute{Required: true},
			"policy_id":       schema.StringAttribute{Required: true},
		},
	}
}

func iamServiceRolePolicyAttachmentPlan(t *testing.T, v iamServiceRolePolicyAttachmentModel) tfsdk.Plan {
	t.Helper()

	s := iamServiceRolePolicyAttachmentSchema()
	attrTypes := map[string]tftypes.Type{
		"id":              tftypes.String,
		"service_role_id": tftypes.String,
		"policy_type":     tftypes.String,
		"policy_id":       tftypes.String,
	}

	return tfsdk.Plan{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"id":              mustTerraformValue(t, v.ID),
				"service_role_id": mustTerraformValue(t, v.ServiceRoleID),
				"policy_type":     mustTerraformValue(t, v.PolicyType),
				"policy_id":       mustTerraformValue(t, v.PolicyID),
			},
		),
	}
}

func iamServiceRolePolicyAttachmentState(t *testing.T, v iamServiceRolePolicyAttachmentModel) tfsdk.State {
	t.Helper()

	s := iamServiceRolePolicyAttachmentSchema()
	attrTypes := map[string]tftypes.Type{
		"id":              tftypes.String,
		"service_role_id": tftypes.String,
		"policy_type":     tftypes.String,
		"policy_id":       tftypes.String,
	}

	return tfsdk.State{
		Schema: s,
		Raw: tftypes.NewValue(
			tftypes.Object{AttributeTypes: attrTypes},
			map[string]tftypes.Value{
				"id":              mustTerraformValue(t, v.ID),
				"service_role_id": mustTerraformValue(t, v.ServiceRoleID),
				"policy_type":     mustTerraformValue(t, v.PolicyType),
				"policy_id":       mustTerraformValue(t, v.PolicyID),
			},
		),
	}
}
