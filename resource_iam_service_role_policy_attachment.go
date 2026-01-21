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

type IamServiceRolePolicyAttachmentResource struct {
	client *APIClient
}

type iamServiceRolePolicyAttachmentModel struct {
	ID            types.String `tfsdk:"id"`
	ServiceRoleID types.String `tfsdk:"service_role_id"`
	PolicyType    types.String `tfsdk:"policy_type"`
	PolicyID      types.String `tfsdk:"policy_id"`
}

func NewIamServiceRolePolicyAttachmentResource() resource.Resource {
	return &IamServiceRolePolicyAttachmentResource{}
}

var _ resource.Resource = (*IamServiceRolePolicyAttachmentResource)(nil)
var _ resource.ResourceWithConfigure = (*IamServiceRolePolicyAttachmentResource)(nil)
var _ resource.ResourceWithImportState = (*IamServiceRolePolicyAttachmentResource)(nil)

func (r *IamServiceRolePolicyAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_service_role_policy_attachment"
}

func (r *IamServiceRolePolicyAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier: {service_role_id}:{policy_type}:{policy_id}.",
			},
			"service_role_id": schema.StringAttribute{
				Required:      true,
				Description:   "Service role resource ID.",
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

func (r *IamServiceRolePolicyAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *IamServiceRolePolicyAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan iamServiceRolePolicyAttachmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceRoleID, ok := requireKnownString(&resp.Diagnostics, plan.ServiceRoleID, path.Root("service_role_id"), "service_role_id")
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

	resp.Diagnostics.Append(r.getPolicy(ctx, policyType, policyID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attachPath := fmt.Sprintf("/service-roles/%s/policies/%s", url.PathEscape(serviceRoleID), url.PathEscape(policyID))
	attachURL := joinURLWithQuery(r.client.iamBaseURL(), attachPath, map[string]string{"policyType": policyType})
	_, status, putDiags := r.client.doJSONInternal(ctx, "PUT", attachURL, nil, nil, false)
	if putDiags.HasError() {
		if status == 405 {
			resp.Diagnostics.Append(r.replaceServiceRolePoliciesAdd(ctx, serviceRoleID, policyType, policyID)...)
			if resp.Diagnostics.HasError() {
				return
			}
		} else {
			resp.Diagnostics.Append(putDiags...)
			return
		}
	}

	plan.ID = types.StringValue(composeAttachmentID(serviceRoleID, policyType, policyID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamServiceRolePolicyAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iamServiceRolePolicyAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceRoleID := state.ServiceRoleID.ValueString()
	policyType := strings.ToLower(state.PolicyType.ValueString())
	policyID := state.PolicyID.ValueString()

	attached, diags := r.isAttached(ctx, serviceRoleID, policyType, policyID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !attached {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(composeAttachmentID(serviceRoleID, policyType, policyID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IamServiceRolePolicyAttachmentResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Unsupported", "All attributes are ForceNew; updates are not supported")
}

func (r *IamServiceRolePolicyAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iamServiceRolePolicyAttachmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceRoleID := state.ServiceRoleID.ValueString()
	policyType := strings.ToLower(state.PolicyType.ValueString())
	policyID := state.PolicyID.ValueString()

	detachPath := fmt.Sprintf("/service-roles/%s/policies/%s", url.PathEscape(serviceRoleID), url.PathEscape(policyID))
	detachURL := joinURLWithQuery(r.client.iamBaseURL(), detachPath, map[string]string{"policyType": policyType})
	_, status, delDiags := r.client.doJSONInternal(ctx, "DELETE", detachURL, nil, nil, false)
	if delDiags.HasError() {
		if status == 405 {
			resp.Diagnostics.Append(r.replaceServiceRolePoliciesRemove(ctx, serviceRoleID, policyType, policyID)...)
			return
		}
		resp.Diagnostics.Append(delDiags...)
		return
	}
}

func (r *IamServiceRolePolicyAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected {service_role_id}:{policy_type}:{policy_id}")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_role_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_type"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *IamServiceRolePolicyAttachmentResource) getPolicy(ctx context.Context, policyType, policyID string) diag.Diagnostics {
	var diags diag.Diagnostics
	var out any
	path := fmt.Sprintf("/policies/%s", url.PathEscape(policyID))
	url := joinURLWithQuery(r.client.iamBaseURL(), path, map[string]string{"policyType": policyType})
	diags.Append(r.client.doJSON(ctx, "GET", url, nil, &out)...)
	return diags
}

func (r *IamServiceRolePolicyAttachmentResource) isAttached(ctx context.Context, serviceRoleID, policyType, policyID string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	policies, diags := r.listServiceRolePoliciesFromServiceRole(ctx, serviceRoleID)
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

func (r *IamServiceRolePolicyAttachmentResource) replaceServiceRolePoliciesAdd(ctx context.Context, serviceRoleID, policyType, policyID string) diag.Diagnostics {
	var diags diag.Diagnostics

	policies, diags := r.listServiceRolePoliciesFromServiceRole(ctx, serviceRoleID)
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
		policies = append(policies, serviceRolePolicyRef{PolicyType: policyType, PolicyResourceId: policyID})
	}

	payload := map[string]any{"serviceRolePolicies": policies}
	postURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/service-roles/%s/policies", url.PathEscape(serviceRoleID)))
	diags.Append(r.client.doJSON(ctx, "POST", postURL, payload, nil)...)
	return diags
}

func (r *IamServiceRolePolicyAttachmentResource) replaceServiceRolePoliciesRemove(ctx context.Context, serviceRoleID, policyType, policyID string) diag.Diagnostics {
	var diags diag.Diagnostics

	policies, diags := r.listServiceRolePoliciesFromServiceRole(ctx, serviceRoleID)
	if diags.HasError() {
		return diags
	}

	filtered := make([]serviceRolePolicyRef, 0, len(policies))
	for _, p := range policies {
		if strings.EqualFold(p.PolicyType, policyType) && p.PolicyResourceId == policyID {
			continue
		}
		filtered = append(filtered, p)
	}

	payload := map[string]any{"serviceRolePolicies": filtered}
	postURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/service-roles/%s/policies", url.PathEscape(serviceRoleID)))
	diags.Append(r.client.doJSON(ctx, "POST", postURL, payload, nil)...)
	return diags
}

type serviceRolePolicyRef struct {
	PolicyType       string `json:"policyType"`
	PolicyResourceId string `json:"policyResourceId"`
}

func (r *IamServiceRolePolicyAttachmentResource) listServiceRolePoliciesFromServiceRole(ctx context.Context, serviceRoleID string) ([]serviceRolePolicyRef, diag.Diagnostics) {
	var diags diag.Diagnostics

	// In some environments, /service-roles/{id}/policies does not surface MANAGED policies for account-owned roles.
	// The service-role GET endpoint with includePolicies is the most reliable source of truth.
	type serviceRolePolicy struct {
		PolicyType string `json:"policyType"`
		ResourceID string `json:"resourceId"`
	}
	type serviceRoleResponse struct {
		ServiceRole struct {
			Policies []serviceRolePolicy `json:"policies"`
		} `json:"serviceRole"`
	}

	var out serviceRoleResponse
	getURL := joinURLWithQuery(
		r.client.iamBaseURL(),
		fmt.Sprintf("/service-roles/%s", url.PathEscape(serviceRoleID)),
		map[string]string{"includePolicies": "true", "resourceOwner": "account"},
	)
	found, getDiags := r.client.doJSONAllowNotFound(ctx, "GET", getURL, nil, &out)
	diags.Append(getDiags...)
	if diags.HasError() {
		return nil, diags
	}
	if !found {
		return []serviceRolePolicyRef{}, diags
	}

	policies := make([]serviceRolePolicyRef, 0, len(out.ServiceRole.Policies))
	for _, p := range out.ServiceRole.Policies {
		if strings.TrimSpace(p.ResourceID) == "" || strings.TrimSpace(p.PolicyType) == "" {
			continue
		}
		policies = append(policies, serviceRolePolicyRef{PolicyType: p.PolicyType, PolicyResourceId: p.ResourceID})
	}

	return policies, diags
}
