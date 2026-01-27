package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

type AuthorizerInstanceResource struct {
	instanceResource
}

func NewAuthorizerInstanceResource() resource.Resource {
	return &AuthorizerInstanceResource{
		instanceResource: instanceResource{
			baseURL: func(client *APIClient) string {
				return client.authorizerBaseURL()
			},
		},
	}
}

var _ resource.Resource = (*AuthorizerInstanceResource)(nil)
var _ resource.ResourceWithConfigure = (*AuthorizerInstanceResource)(nil)
var _ resource.ResourceWithImportState = (*AuthorizerInstanceResource)(nil)

func (r *AuthorizerInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authorizer_instance"
}

func (r *AuthorizerInstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Authorizer instance resource ID. Immutable. If omitted, the provider will generate one.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human readable instance name.",
			},
			"configuration_resource_id": schema.StringAttribute{
				Optional:    true,
				Description: "Authorizer configuration resource ID to apply to this instance.",
			},
			"inline_configuration": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inline authorizer configuration JSON (string). If omitted, the server may default this to an empty object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint": instanceEndpointSchemaAttribute(),
		},
	}
}
