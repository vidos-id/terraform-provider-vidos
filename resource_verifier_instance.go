package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

type VerifierInstanceResource struct {
	instanceResource
}

func NewVerifierInstanceResource() resource.Resource {
	return &VerifierInstanceResource{
		instanceResource: instanceResource{
			baseURL: func(client *APIClient) string {
				return client.verifierBaseURL()
			},
		},
	}
}

var _ resource.Resource = (*VerifierInstanceResource)(nil)
var _ resource.ResourceWithConfigure = (*VerifierInstanceResource)(nil)
var _ resource.ResourceWithImportState = (*VerifierInstanceResource)(nil)

func (r *VerifierInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_verifier_instance"
}

func (r *VerifierInstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Verifier instance resource ID. Immutable. If omitted, the provider will generate one.",
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
				Description: "Verifier configuration resource ID to apply to this instance.",
			},
			"inline_configuration": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inline verifier configuration JSON (string). If omitted, the server may default this to an empty object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
