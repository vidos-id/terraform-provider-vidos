package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

type ResolverInstanceResource struct {
	instanceResource
}

func NewResolverInstanceResource() resource.Resource {
	return &ResolverInstanceResource{
		instanceResource: instanceResource{
			baseURL: func(client *APIClient) string {
				return client.resolverBaseURL()
			},
		},
	}
}

var _ resource.Resource = (*ResolverInstanceResource)(nil)
var _ resource.ResourceWithConfigure = (*ResolverInstanceResource)(nil)
var _ resource.ResourceWithImportState = (*ResolverInstanceResource)(nil)

func (r *ResolverInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resolver_instance"
}

func (r *ResolverInstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Resolver instance resource ID. Immutable. If omitted, the provider will generate one.",
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
				Description: "Resolver configuration resource ID to apply to this instance.",
			},
			"inline_configuration": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inline resolver configuration JSON (string). If omitted, the server may default this to an empty object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint": instanceEndpointSchemaAttribute(),
		},
	}
}
