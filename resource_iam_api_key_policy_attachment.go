package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IamApiKeyPolicyAttachmentResource struct {
	client *APIClient
}

type iamApiKeyPolicyAttachmentModel struct {
	ID         types.String `tfsdk:"id"`
	ApiKeyID   types.String `tfsdk:"api_key_id"`
	PolicyType types.String `tfsdk:"policy_type"`
	PolicyID   types.String `tfsdk:"policy_id"`
}

func NewIamApiKeyPolicyAttachmentResource() resource.Resource {
	return &IamApiKeyPolicyAttachmentResource{}
}

var _ resource.Resource = (*IamApiKeyPolicyAttachmentResource)(nil)
var _ resource.ResourceWithConfigure = (*IamApiKeyPolicyAttachmentResource)(nil)
var _ resource.ResourceWithImportState = (*IamApiKeyPolicyAttachmentResource)(nil)

func (r *IamApiKeyPolicyAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_api_key_policy_attachment"
}

func (r *IamApiKeyPolicyAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier: {api_key_id}:{policy_type}:{policy_id}.",
			},
			"api_key_id": schema.StringAttribute{
				Required:      true,
				Description:   "API key resource ID (32 hex).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"policy_type": schema.StringAttribute{
				Required:      true,
				Description:   "Policy type. Can be account or managed.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"policy_id": schema.StringAttribute{
				Required:      true,
				Description:   "Policy resource ID.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *IamApiKeyPolicyAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *IamApiKeyPolicyAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan iamApiKeyPolicyAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKeyID, ok := requireKnownString(&resp.Diagnostics, plan.ApiKeyID, path.Root("api_key_id"), "api_key_id")
	if !ok {
		return
	}
	policyType, ok := requireKnownString(&resp.Diagnostics, plan.PolicyType, path.Root("policy_type"), "policy_type")
	if !ok {
		return
	}
	policyID, ok := requireKnownString(&resp.Diagnostics, plan.PolicyID, path.Root("policy_id"), "policy_id")
	if !ok {
		return
	}

	policyType = strings.ToLower(policyType)
	if policyType != "account" && policyType != "managed" {
		resp.Diagnostics.AddAttributeError(path.Root("policy_type"), "Invalid policy_type", "policy_type must be account or managed")
		return
	}

	// Fail-fast: verify policy exists before attaching.
	resp.Diagnostics.Append(r.getPolicy(ctx, policyType, policyID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attachPath := fmt.Sprintf("/api-keys/%s/policies/%s", url.PathEscape(apiKeyID), url.PathEscape(policyID))
	attachURL := joinURLWithQuery(r.client.iamBaseURL(), attachPath, map[string]string{"policyType": policyType})
	_, status, putDiags := r.client.doJSONInternal(ctx, "PUT", attachURL, nil, nil, false)
	if putDiags.HasError() {
		if status == 405 {
			resp.Diagnostics.Append(r.replaceApiKeyPoliciesAdd(ctx, apiKeyID, policyType, policyID)...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			resp.Diagnostics.Append(putDiags...)
			return
		}
	}

	plan.ID = types.StringValue(composeAttachmentID(apiKeyID, policyType, policyID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamApiKeyPolicyAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iamApiKeyPolicyAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKeyID := state.ApiKeyID.ValueString()
	policyType := strings.ToLower(state.PolicyType.ValueString())
	policyID := state.PolicyID.ValueString()

	attached, diags := r.isAttached(ctx, apiKeyID, policyType, policyID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !attached {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(composeAttachmentID(apiKeyID, policyType, policyID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IamApiKeyPolicyAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported", "All attributes are ForceNew; updates are not supported")
}

func (r *IamApiKeyPolicyAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iamApiKeyPolicyAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKeyID := state.ApiKeyID.ValueString()
	policyType := strings.ToLower(state.PolicyType.ValueString())
	policyID := state.PolicyID.ValueString()

	detachPath := fmt.Sprintf("/api-keys/%s/policies/%s", url.PathEscape(apiKeyID), url.PathEscape(policyID))
	detachURL := joinURLWithQuery(r.client.iamBaseURL(), detachPath, map[string]string{"policyType": policyType})
	_, status, delDiags := r.client.doJSONInternal(ctx, "DELETE", detachURL, nil, nil, false)
	if delDiags.HasError() {
		if status == 405 {
			resp.Diagnostics.Append(r.replaceApiKeyPoliciesRemove(ctx, apiKeyID, policyType, policyID)...)
			return
		}
		resp.Diagnostics.Append(delDiags...)
		return
	}
}

func (r *IamApiKeyPolicyAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected {api_key_id}:{policy_type}:{policy_id}")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("api_key_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_type"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *IamApiKeyPolicyAttachmentResource) getPolicy(ctx context.Context, policyType, policyID string) diag.Diagnostics {
	var diags diag.Diagnostics
	var out any
	path := fmt.Sprintf("/policies/%s", url.PathEscape(policyID))
	url := joinURLWithQuery(r.client.iamBaseURL(), path, map[string]string{"policyType": policyType})
	diags.Append(r.client.doJSON(ctx, "GET", url, nil, &out)...)
	return diags
}

func (r *IamApiKeyPolicyAttachmentResource) isAttached(ctx context.Context, apiKeyID, policyType, policyID string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	policies, diags := r.listApiKeyPolicies(ctx, apiKeyID)
	if diags.HasError() {
		return false, diags
	}
	for _, p := range policies {
		if strings.EqualFold(p.PolicyType, policyType) && p.PolicyResourceId == policyID {
			return true, diags
		}
	}
	return false, diags
}

func (r *IamApiKeyPolicyAttachmentResource) replaceApiKeyPoliciesAdd(ctx context.Context, apiKeyID, policyType, policyID string) diag.Diagnostics {
	var diags diag.Diagnostics

	policies, diags := r.listApiKeyPolicies(ctx, apiKeyID)
	if diags.HasError() {
		return diags
	}

	exists := false
	for _, p := range policies {
		if strings.EqualFold(p.PolicyType, policyType) && p.PolicyResourceId == policyID {
			exists = true
			break
		}
	}
	if !exists {
		policies = append(policies, apiKeyPolicyRef{PolicyType: policyType, PolicyResourceId: policyID})
	}

	payload := map[string]any{"apiKeyPolicies": policies}
	postURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/api-keys/%s/policies", url.PathEscape(apiKeyID)))
	diags.Append(r.client.doJSON(ctx, "POST", postURL, payload, nil)...)
	return diags
}

func (r *IamApiKeyPolicyAttachmentResource) replaceApiKeyPoliciesRemove(ctx context.Context, apiKeyID, policyType, policyID string) diag.Diagnostics {
	var diags diag.Diagnostics

	policies, diags := r.listApiKeyPolicies(ctx, apiKeyID)
	if diags.HasError() {
		return diags
	}

	filtered := make([]apiKeyPolicyRef, 0, len(policies))
	for _, p := range policies {
		if strings.EqualFold(p.PolicyType, policyType) && p.PolicyResourceId == policyID {
			continue
		}
		filtered = append(filtered, p)
	}

	payload := map[string]any{"apiKeyPolicies": filtered}
	postURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/api-keys/%s/policies", url.PathEscape(apiKeyID)))
	diags.Append(r.client.doJSON(ctx, "POST", postURL, payload, nil)...)
	return diags
}

type apiKeyPolicyRef struct {
	PolicyType       string `json:"policyType"`
	PolicyResourceId string `json:"policyResourceId"`
}

func (r *IamApiKeyPolicyAttachmentResource) listApiKeyPolicies(ctx context.Context, apiKeyID string) ([]apiKeyPolicyRef, diag.Diagnostics) {
	var diags diag.Diagnostics

	type policiesResponse struct {
		ApiKeyPolicies []apiKeyPolicyRef `json:"apiKeyPolicies"`
	}

	var out policiesResponse
	path := fmt.Sprintf("/api-keys/%s/policies", url.PathEscape(apiKeyID))
	diags.Append(r.client.doJSON(ctx, "GET", joinURL(r.client.iamBaseURL(), path), nil, &out)...)
	if diags.HasError() {
		return nil, diags
	}
	return out.ApiKeyPolicies, diags
}

func composeAttachmentID(principalID, policyType, policyID string) string {
	return principalID + ":" + strings.ToLower(policyType) + ":" + policyID
}
