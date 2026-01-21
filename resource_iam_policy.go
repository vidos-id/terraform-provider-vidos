package main

import (
	"context"
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

type IamPolicyResource struct {
	client *APIClient
}

type iamPolicyModel struct {
	ResourceID types.String `tfsdk:"resource_id"`
	Name       types.String `tfsdk:"name"`
	Document   types.String `tfsdk:"document"`
}

func NewIamPolicyResource() resource.Resource {
	return &IamPolicyResource{}
}

var _ resource.Resource = (*IamPolicyResource)(nil)
var _ resource.ResourceWithConfigure = (*IamPolicyResource)(nil)
var _ resource.ResourceWithImportState = (*IamPolicyResource)(nil)

func (r *IamPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_policy"
}

func (r *IamPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "Policy resource ID. This is the policy identifier in Vidos IAM and is immutable. " +
					"Only account policies are managed by this provider. If omitted, the provider will generate one.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human readable policy name.",
			},
			"document": schema.StringAttribute{
				Required:    true,
				Description: "Policy document JSON (string).",
			},
		},
	}
}

func (r *IamPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *IamPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config iamPolicyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan iamPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, ok := resolveOrGenerateResourceID(&resp.Diagnostics, config.ResourceID, path.Root("resource_id"))
	if !ok {
		return
	}

	document := parseJSONToAny(&resp.Diagnostics, plan.Document.ValueString(), path.Root("document"), "document")
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"policyResourceId": resourceID,
		"policy": map[string]any{
			"name":     plan.Name.ValueString(),
			"document": document,
		},
	}

	var out any
	resp.Diagnostics.Append(r.client.doJSON(ctx, "POST", joinURL(r.client.iamBaseURL(), "/policies"), payload, &out)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found, diags := r.readIntoState(ctx, resourceID, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.Diagnostics.AddError("API error", "Policy was created but could not be read back")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iamPolicyModel
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

func (r *IamPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan iamPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := plan.ResourceID.ValueString()
	document := parseJSONToAny(&resp.Diagnostics, plan.Document.ValueString(), path.Root("document"), "document")
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"policy": map[string]any{
			"name":     plan.Name.ValueString(),
			"document": document,
		},
	}

	resp.Diagnostics.Append(r.client.doJSON(ctx, "PUT", joinURL(r.client.iamBaseURL(), fmt.Sprintf("/policies/%s", url.PathEscape(resourceID))), payload, nil)...)
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

func (r *IamPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iamPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.ResourceID.ValueString()
	_, delDiags := r.client.doJSONAllowNotFound(ctx, "DELETE", joinURL(r.client.iamBaseURL(), fmt.Sprintf("/policies/%s", url.PathEscape(resourceID))), nil, nil)
	resp.Diagnostics.Append(delDiags...)
}

func (r *IamPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}

func (r *IamPolicyResource) readIntoState(ctx context.Context, resourceID string, state *iamPolicyModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	type policyResponse struct {
		Policy struct {
			ResourceID string `json:"resourceId"`
			Name       string `json:"name"`
			Document   any    `json:"document"`
			PolicyType string `json:"policyType"`
		} `json:"policy"`
	}

	var out policyResponse
	getURL := joinURLWithQuery(r.client.iamBaseURL(), fmt.Sprintf("/policies/%s", url.PathEscape(resourceID)), map[string]string{"policyType": "account"})
	found, getDiags := r.client.doJSONAllowNotFound(ctx, "GET", getURL, nil, &out)
	diags.Append(getDiags...)
	if diags.HasError() {
		return false, diags
	}
	if !found {
		return false, diags
	}
	if out.Policy.PolicyType != "" && out.Policy.PolicyType != "account" {
		diags.AddError("Unsupported policy type", "vidos_iam_policy manages account policies only")
		return true, diags
	}

	doc, err := jsonMarshal(out.Policy.Document)
	if err != nil {
		diags.AddError("Document encode error", err.Error())
		return true, diags
	}

	state.ResourceID = types.StringValue(out.Policy.ResourceID)
	state.Name = types.StringValue(out.Policy.Name)
	state.Document = types.StringValue(string(doc))

	return true, diags
}
