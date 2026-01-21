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

type IamApiKeyResource struct {
	client *APIClient
}

type iamApiKeyModel struct {
	ResourceID           types.String `tfsdk:"resource_id"`
	Name                 types.String `tfsdk:"name"`
	InlinePolicyDocument types.String `tfsdk:"inline_policy_document"`
	ApiSecret            types.String `tfsdk:"api_secret"`
}

func NewIamApiKeyResource() resource.Resource {
	return &IamApiKeyResource{}
}

var _ resource.Resource = (*IamApiKeyResource)(nil)
var _ resource.ResourceWithConfigure = (*IamApiKeyResource)(nil)
var _ resource.ResourceWithImportState = (*IamApiKeyResource)(nil)

func (r *IamApiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iam_api_key"
}

func (r *IamApiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"resource_id": schema.StringAttribute{
				Computed:    true,
				Description: "API key resource ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human readable API key name.",
			},
			"inline_policy_document": schema.StringAttribute{
				Optional:    true,
				Description: "Inline policy document JSON (string) for this API key.",
			},
			"api_secret": schema.StringAttribute{
				Computed:      true,
				Sensitive:     true,
				Description:   "API key secret (write-only). Returned only on create; not retrievable and will remain unknown after import.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *IamApiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *IamApiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan iamApiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := map[string]any{
		"name": plan.Name.ValueString(),
	}
	if !plan.InlinePolicyDocument.IsNull() && !plan.InlinePolicyDocument.IsUnknown() {
		apiKey["inlinePolicyDocument"] = parseJSONToAny(&resp.Diagnostics, plan.InlinePolicyDocument.ValueString(), path.Root("inline_policy_document"), "inline_policy_document")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	payload := map[string]any{"apiKey": apiKey}

	type apiKeyResponse struct {
		ApiKey struct {
			ResourceID           string `json:"resourceId"`
			Name                 string `json:"name"`
			InlinePolicyDocument any    `json:"inlinePolicyDocument"`
			ApiSecret            string `json:"apiSecret"`
		} `json:"apiKey"`
	}

	var out apiKeyResponse
	resp.Diagnostics.Append(r.client.doJSON(ctx, "POST", joinURL(r.client.iamBaseURL(), "/api-keys"), payload, &out)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ResourceID = types.StringValue(out.ApiKey.ResourceID)
	plan.Name = types.StringValue(out.ApiKey.Name)
	if out.ApiKey.InlinePolicyDocument == nil {
		plan.InlinePolicyDocument = types.StringNull()
	} else {
		b, err := json.Marshal(out.ApiKey.InlinePolicyDocument)
		if err != nil {
			resp.Diagnostics.AddError("inlinePolicyDocument encode error", err.Error())
			return
		}
		plan.InlinePolicyDocument = types.StringValue(string(b))
	}
	if out.ApiKey.ApiSecret == "" {
		plan.ApiSecret = types.StringUnknown()
	} else {
		plan.ApiSecret = types.StringValue(out.ApiKey.ApiSecret)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamApiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state iamApiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	existingSecret := state.ApiSecret

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

	state.ApiSecret = existingSecret
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IamApiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan iamApiKeyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, ok := requireKnownString(&resp.Diagnostics, plan.ResourceID, path.Root("resource_id"), "resource_id")
	if !ok {
		return
	}

	apiKey := map[string]any{"name": plan.Name.ValueString()}
	if plan.InlinePolicyDocument.IsNull() {
		apiKey["inlinePolicyDocument"] = nil
	} else if !plan.InlinePolicyDocument.IsUnknown() {
		apiKey["inlinePolicyDocument"] = parseJSONToAny(&resp.Diagnostics, plan.InlinePolicyDocument.ValueString(), path.Root("inline_policy_document"), "inline_policy_document")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	payload := map[string]any{"apiKey": apiKey}

	updateURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/api-keys/%s", url.PathEscape(resourceID)))
	resp.Diagnostics.Append(r.client.doJSON(ctx, "POST", updateURL, payload, nil)...)
	if resp.Diagnostics.HasError() {
		return
	}

	existingSecret := plan.ApiSecret
	found, diags := r.readIntoState(ctx, resourceID, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	plan.ApiSecret = existingSecret

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IamApiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state iamApiKeyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.ResourceID.ValueString()
	delURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/api-keys/%s", url.PathEscape(resourceID)))
	_, delDiags := r.client.doJSONAllowNotFound(ctx, "DELETE", delURL, nil, nil)
	resp.Diagnostics.Append(delDiags...)
}

func (r *IamApiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}

func (r *IamApiKeyResource) readIntoState(ctx context.Context, resourceID string, state *iamApiKeyModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	type apiKeyResponse struct {
		ApiKey struct {
			ResourceID           string `json:"resourceId"`
			Name                 string `json:"name"`
			InlinePolicyDocument any    `json:"inlinePolicyDocument"`
		} `json:"apiKey"`
	}

	var out apiKeyResponse
	getURL := joinURL(r.client.iamBaseURL(), fmt.Sprintf("/api-keys/%s", url.PathEscape(resourceID)))
	found, getDiags := r.client.doJSONAllowNotFound(ctx, "GET", getURL, nil, &out)
	diags.Append(getDiags...)
	if diags.HasError() {
		return false, diags
	}
	if !found {
		return false, diags
	}

	state.ResourceID = types.StringValue(out.ApiKey.ResourceID)
	state.Name = types.StringValue(out.ApiKey.Name)
	if out.ApiKey.InlinePolicyDocument == nil {
		state.InlinePolicyDocument = types.StringNull()
	} else {
		b, err := json.Marshal(out.ApiKey.InlinePolicyDocument)
		if err != nil {
			diags.AddError("inlinePolicyDocument encode error", err.Error())
			return true, diags
		}
		state.InlinePolicyDocument = types.StringValue(string(b))
	}

	return true, diags
}
