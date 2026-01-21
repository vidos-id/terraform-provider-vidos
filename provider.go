package main

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type VidosProvider struct {
	version string
}

type providerModel struct {
	Region types.String `tfsdk:"region"`
	ApiKey types.String `tfsdk:"api_key"`
}

type providerConfig struct {
	domain        string
	defaultRegion string
	apiKeySecret  string
}

func New() provider.Provider {
	return &VidosProvider{version: version}
}

func (p *VidosProvider) Metadata(_ context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vidos"
	resp.Version = p.version
}

func (p *VidosProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Default region segment for service management endpoints (e.g. eu). IAM is always global.",
				Validators:  []validator.String{regionValidator{}},
			},
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Vidos IAM API secret (64 hex) used as Authorization: Bearer <api_key>.",
			},
		},
	}
}

func (p *VidosProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg, diags := buildProviderConfig(config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := NewAPIClient(cfg)
	resp.ResourceData = client
	resp.DataSourceData = client

	tflog.Info(ctx, "Configured Vidos provider", map[string]any{
		"region": cfg.defaultRegion,
	})
}

func buildProviderConfig(config providerModel) (providerConfig, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Domain is set at build time via ldflags
	domain := defaultDomain

	defaultRegion := getFirstNonEmpty(config.Region, os.Getenv("VIDOS_REGION"))
	if defaultRegion == "" {
		defaultRegion = "eu"
	}

	apiKey := getFirstNonEmpty(config.ApiKey, os.Getenv("VIDOS_API_KEY"))
	if apiKey == "" {
		diags.AddError(
			"Missing API key",
			"Set provider attribute api_key or env var VIDOS_API_KEY.",
		)
		return providerConfig{}, diags
	}

	return providerConfig{
		domain:        domain,
		defaultRegion: defaultRegion,
		apiKeySecret:  apiKey,
	}, diags
}

func getFirstNonEmpty(attr types.String, env string) string {
	if !attr.IsNull() && !attr.IsUnknown() {
		return strings.TrimSpace(attr.ValueString())
	}
	return strings.TrimSpace(env)
}

func buildManagementBaseURL(service, region, domain string) string {
	return "https://" + service + ".management." + region + "." + domain
}

func (p *VidosProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewIamApiKeyResource,
		NewIamPolicyResource,
		NewIamApiKeyPolicyAttachmentResource,
		NewIamServiceRoleResource,
		NewIamServiceRolePolicyAttachmentResource,
		NewResolverConfigurationResource,
		NewResolverInstanceResource,
		NewVerifierConfigurationResource,
		NewVerifierInstanceResource,
		NewValidatorConfigurationResource,
		NewValidatorInstanceResource,
		NewAuthorizerConfigurationResource,
		NewAuthorizerInstanceResource,
		NewGatewayConfigurationResource,
		NewGatewayInstanceResource,
	}
}

func (p *VidosProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

var _ provider.Provider = (*VidosProvider)(nil)

// Helper to add a nice error when required config is unknown.
func requireKnownString(diags *diag.Diagnostics, val types.String, attrPath path.Path, name string) (string, bool) {
	if val.IsUnknown() {
		diags.AddAttributeError(attrPath, "Unknown value", name+" must be known during planning.")
		return "", false
	}
	if val.IsNull() {
		diags.AddAttributeError(attrPath, "Missing value", name+" is required.")
		return "", false
	}
	out := strings.TrimSpace(val.ValueString())
	if out == "" {
		diags.AddAttributeError(attrPath, "Invalid value", name+" must not be empty.")
		return "", false
	}
	return out, true
}
