package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

type ValidatorInstanceResource struct {
	instanceResource
}

func NewValidatorInstanceResource() resource.Resource {
	return &ValidatorInstanceResource{
		instanceResource: instanceResource{
			baseURL: func(client *APIClient) string {
				return client.validatorBaseURL()
			},
		},
	}
}

var _ resource.Resource = (*ValidatorInstanceResource)(nil)
var _ resource.ResourceWithConfigure = (*ValidatorInstanceResource)(nil)
var _ resource.ResourceWithImportState = (*ValidatorInstanceResource)(nil)

func (r *ValidatorInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_validator_instance"
}

func (r *ValidatorInstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Validator instance resource ID. Immutable. If omitted, the provider will generate one.",
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
				Description: "Validator configuration resource ID to apply to this instance.",
			},
			"inline_configuration": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inline validator configuration JSON (string). If omitted, the server may default this to an empty object.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint": instanceEndpointSchemaAttribute(),
		},
	}
}
