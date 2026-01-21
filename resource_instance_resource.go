package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type instanceResource struct {
	client  *APIClient
	baseURL func(*APIClient) string
}

// Note: this is an embedded helper; the wrapper resources implement Metadata/Schema.

func (r *instanceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*APIClient)
}

func (r *instanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var config instanceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan instanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID, ok := resolveOrGenerateResourceID(&resp.Diagnostics, config.ResourceID, path.Root("resource_id"))
	if !ok {
		return
	}

	instance := map[string]any{
		"name": plan.Name.ValueString(),
	}
	if plan.ConfigurationResourceID.IsNull() {
		instance["configurationResourceId"] = nil
	} else if !plan.ConfigurationResourceID.IsUnknown() {
		instance["configurationResourceId"] = plan.ConfigurationResourceID.ValueString()
	}
	if !plan.InlineConfiguration.IsNull() && !plan.InlineConfiguration.IsUnknown() {
		instance["inlineConfiguration"] = parseJSONToAny(&resp.Diagnostics, plan.InlineConfiguration.ValueString(), path.Root("inline_configuration"), "inline_configuration")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	payload := instanceCreatePayload(resourceID, instance)
	resp.Diagnostics.Append(createInstance(ctx, r.client, r.baseURL(r.client), payload)...)
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

func (r *instanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceModel
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

func (r *instanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan instanceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := plan.ResourceID.ValueString()
	instance := map[string]any{
		"name": plan.Name.ValueString(),
	}
	if plan.ConfigurationResourceID.IsNull() {
		instance["configurationResourceId"] = nil
	} else if !plan.ConfigurationResourceID.IsUnknown() {
		instance["configurationResourceId"] = plan.ConfigurationResourceID.ValueString()
	}
	if plan.InlineConfiguration.IsNull() {
		// omit to keep server default when unset
	} else if !plan.InlineConfiguration.IsUnknown() {
		instance["inlineConfiguration"] = parseJSONToAny(&resp.Diagnostics, plan.InlineConfiguration.ValueString(), path.Root("inline_configuration"), "inline_configuration")
		if resp.Diagnostics.HasError() {
			return
		}
	}

	payload := instanceUpdatePayload(instance)
	resp.Diagnostics.Append(updateInstance(ctx, r.client, r.baseURL(r.client), resourceID, payload)...)
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

func (r *instanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resourceID := state.ResourceID.ValueString()
	resp.Diagnostics.Append(deleteInstance(ctx, r.client, r.baseURL(r.client), resourceID)...)
}

func (r *instanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_id"), req, resp)
}

func (r *instanceResource) readIntoState(ctx context.Context, resourceID string, state *instanceModel) (bool, diag.Diagnostics) {
	found, out, inlineJSON, diags := readInstanceIntoState(ctx, r.client, r.baseURL(r.client), resourceID)
	if diags.HasError() {
		return false, diags
	}
	if !found {
		return false, diags
	}

	state.ResourceID = types.StringValue(out.Instance.ResourceID)
	state.Name = types.StringValue(out.Instance.Name)
	if out.Instance.ConfigurationResourceID == "" {
		state.ConfigurationResourceID = types.StringNull()
	} else {
		state.ConfigurationResourceID = types.StringValue(out.Instance.ConfigurationResourceID)
	}
	if out.Instance.InlineConfiguration == nil {
		state.InlineConfiguration = types.StringNull()
	} else {
		state.InlineConfiguration = types.StringValue(inlineJSON)
	}

	return true, diags
}
