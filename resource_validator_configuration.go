package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ValidatorConfigurationResource struct {
	client *APIClient
}

func NewValidatorConfigurationResource() resource.Resource {
	return &ValidatorConfigurationResource{}
}

var _ resource.Resource = (*ValidatorConfigurationResource)(nil)
var _ resource.ResourceWithConfigure = (*ValidatorConfigurationResource)(nil)
var _ resource.ResourceWithImportState = (*ValidatorConfigurationResource)(nil)

func (r *ValidatorConfigurationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_validator_configuration"
}

func (r *ValidatorConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Validator configuration resource ID. Immutable. If omitted, the provider will generate one.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human readable configuration name.",
			},
			"values": schema.StringAttribute{
				Required:    true,
				Description: "Validator configuration values JSON (string).",
			},
		},
	}
}

func (r *ValidatorConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *ValidatorConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config configurationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan configurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, ok := resolveOrGenerateResourceID(&resp.Diagnostics, config.ResourceID, path.Root("resource_id"))
	if !ok {
		return
	}

	values := parseJSONToAny(&resp.Diagnostics, plan.Values.ValueString(), path.Root("values"), "values")
	if resp.Diagnostics.HasError() {
		return
	}

	payload := configurationCreatePayload(resourceID, plan.Name.ValueString(), values)
	resp.Diagnostics.Append(createConfiguration(ctx, r.client, r.client.validatorBaseURL(), payload)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, diags := r.readIntoState(ctx, resourceID, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ValidatorConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state configurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.ResourceID.ValueString()
	found, diags := r.readIntoState(ctx, resourceID, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ValidatorConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan configurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := plan.ResourceID.ValueString()
	values := parseJSONToAny(&resp.Diagnostics, plan.Values.ValueString(), path.Root("values"), "values")
	if resp.Diagnostics.HasError() {
		return
	}
	payload := configurationUpdatePayload(plan.Name.ValueString(), values)
	resp.Diagnostics.Append(updateConfiguration(ctx, r.client, r.client.validatorBaseURL(), resourceID, payload)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, diags := r.readIntoState(ctx, resourceID, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ValidatorConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state configurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.ResourceID.ValueString()
	resp.Diagnostics.Append(deleteConfiguration(ctx, r.client, r.client.validatorBaseURL(), resourceID)...)
}

func (r *ValidatorConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}

func (r *ValidatorConfigurationResource) readIntoState(ctx context.Context, resourceID string, state *configurationModel) (bool, diag.Diagnostics) {
	found, out, valuesJSON, diags := readConfigurationIntoState(ctx, r.client, r.client.validatorBaseURL(), resourceID)
	if diags.HasError() {
		return false, diags
	}
	if !found {
		return false, diags
	}

	state.ResourceID = types.StringValue(out.Configuration.ResourceID)
	state.Name = types.StringValue(out.Configuration.Name)
	state.Values = types.StringValue(valuesJSON)

	return true, diags
}
