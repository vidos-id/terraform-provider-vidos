package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

type GatewayInstanceResource struct {
	instanceResource
}

func NewGatewayInstanceResource() resource.Resource {
	return &GatewayInstanceResource{
		instanceResource: instanceResource{
			baseURL: func(client *APIClient) string {
				return client.gatewayBaseURL()
			},
		},
	}
}

var _ resource.Resource = (*GatewayInstanceResource)(nil)
var _ resource.ResourceWithConfigure = (*GatewayInstanceResource)(nil)
var _ resource.ResourceWithImportState = (*GatewayInstanceResource)(nil)

func (r *GatewayInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway_instance"
}

func (r *GatewayInstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Gateway instance resource ID. Immutable. If omitted, the provider will generate one.",
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
				Description: "Gateway configuration resource ID to apply to this instance.",
			},
			"inline_configuration": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inline gateway configuration JSON (string). If omitted, the server may default this to an empty object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
