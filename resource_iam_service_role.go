package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IamServiceRoleResource struct {
	client *APIClient
}

type iamServiceRoleModel struct {
	ResourceID           types.String `tfsdk:"resource_id"`
	Name                 types.String `tfsdk:"name"`
	InlinePolicyDocument types.String `tfsdk:"inline_policy_document"`
}

func NewIamServiceRoleResource() resource.Resource {
	return &IamServiceRoleResource{}
}

var _ resource.Resource = (*IamServiceRoleResource)(nil)
var _ resource.ResourceWithConfigure = (*IamServiceRoleResource)(nil)
var _ resource.ResourceWithImportState = (*IamServiceRoleResource)(nil)

func (r *IamServiceRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_service_role"
}

func (r *IamServiceRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Service role resource ID. Immutable. If omitted, the provider will generate one.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human readable service role name.",
			},
			"inline_policy_document": schema.StringAttribute{
				Optional:    true,
				Description: "Inline policy document JSON (string) for this service role.",
			},
		},
	}
}

func (r *IamServiceRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *IamServiceRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config iamServiceRoleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan iamServiceRoleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, ok := resolveOrGenerateResourceID(&resp.Diagnostics, config.ResourceID, path.Root("resource_id"))
	if !ok {
		return
	}

	serviceRole := map[string]any{"name": plan.Name.ValueString()}
	if !plan.InlinePolicyDocument.IsNull() && !plan.InlinePolicyDocument.IsUnknown() {
		serviceRole["inlinePolicyDocument"] = parseJSONToAny(&resp.Diagnostics, plan.InlinePolicyDocument.ValueString(), path.Root("inline_policy_document"), "inline_policy_document")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	payload := map[string]any{
		"serviceRoleResourceId": resourceID,
		"serviceRole":           serviceRole,
	}

	var out any
	resp.Diagnostics.Append(r.client.doJSON(ctx, "POST", joinURL(r.client.iamBaseURL(), "/service-roles"), payload, &out)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.readIntoState(ctx, resourceID, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("API error", "Service role was created but could not be read back")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamServiceRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iamServiceRoleModel
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

func (r *IamServiceRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan iamServiceRoleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := plan.ResourceID.ValueString()

	serviceRole := map[string]any{"name": plan.Name.ValueString()}
	if plan.InlinePolicyDocument.IsNull() {
		serviceRole["inlinePolicyDocument"] = nil
	} else if !plan.InlinePolicyDocument.IsUnknown() {
		serviceRole["inlinePolicyDocument"] = parseJSONToAny(&resp.Diagnostics, plan.InlinePolicyDocument.ValueString(), path.Root("inline_policy_document"), "inline_policy_document")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	payload := map[string]any{"serviceRole": serviceRole}
	updateURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/service-roles/%s", url.PathEscape(resourceID)))
	resp.Diagnostics.Append(r.client.doJSON(ctx, "POST", updateURL, payload, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.readIntoState(ctx, resourceID, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamServiceRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iamServiceRoleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.ResourceID.ValueString()
	delURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/service-roles/%s", url.PathEscape(resourceID)))
	_, delDiags := r.client.doJSONAllowNotFound(ctx, "DELETE", delURL, nil, nil)
	resp.Diagnostics.Append(delDiags...)
}

func (r *IamServiceRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}

func (r *IamServiceRoleResource) readIntoState(ctx context.Context, resourceID string, state *iamServiceRoleModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	type serviceRoleResponse struct {
		ServiceRole struct {
			ResourceID           string `json:"resourceId"`
			Name                 string `json:"name"`
			InlinePolicyDocument any    `json:"inlinePolicyDocument"`
		} `json:"serviceRole"`
	}

	var out serviceRoleResponse
	getURL := joinURLWithQuery(r.client.iamBaseURL(), fmt.Sprintf("/service-roles/%s", url.PathEscape(resourceID)), map[string]string{"resourceOwner": "account"})
	found, getDiags := r.client.doJSONAllowNotFound(ctx, "GET", getURL, nil, &out)
	diags.Append(getDiags...)
	if diags.HasError() {
		return false, diags
	}
	if !found {
		return false, diags
	}

	state.ResourceID = types.StringValue(out.ServiceRole.ResourceID)
	state.Name = types.StringValue(out.ServiceRole.Name)
	if out.ServiceRole.InlinePolicyDocument == nil {
		state.InlinePolicyDocument = types.StringNull()
	} else {
		b, err := json.Marshal(out.ServiceRole.InlinePolicyDocument)
		if err != nil {
			diags.AddError("inlinePolicyDocument encode error", err.Error())
			return true, diags
		}
		state.InlinePolicyDocument = types.StringValue(string(b))
	}

	return true, diags
}
