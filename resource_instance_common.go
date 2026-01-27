package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// instanceSleepFn exists to make deletion-wait behavior unit-testable.
// Production code uses time.Sleep.
var instanceSleepFn = time.Sleep

// instanceNowFn exists to make retry timing deterministic in tests.
// Production code uses time.Now.
var instanceNowFn = time.Now

type instanceModel struct {
	ResourceID              types.String `tfsdk:"resource_id"`
	Name                    types.String `tfsdk:"name"`
	ConfigurationResourceID types.String `tfsdk:"configuration_resource_id"`
	InlineConfiguration     types.String `tfsdk:"inline_configuration"`
	Endpoint                types.String `tfsdk:"endpoint"`
}

type instanceReadResponse struct {
	Instance struct {
		ResourceID              string `json:"resourceId"`
		Name                    string `json:"name"`
		ConfigurationResourceID string `json:"configurationResourceId"`
		InlineConfiguration     any    `json:"inlineConfiguration"`
		Endpoint                string `json:"endpoint"`
	} `json:"instance"`
}

func instanceEndpointToState(endpoint string) types.String {
	if endpoint == "" {
		return types.StringNull()
	}
	return types.StringValue(endpoint)
}

func instanceEndpointSchemaAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:    true,
		Sensitive:   false,
		Description: "Platform-reported endpoint (pass-through).",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func instanceCreatePayload(resourceID string, instance map[string]any) map[string]any {
	return map[string]any{
		"instanceResourceId": resourceID,
		"instance":           instance,
	}
}

func instanceUpdatePayload(instance map[string]any) map[string]any {
	return map[string]any{
		"instance": instance,
	}
}

func createInstance(ctx context.Context, client *APIClient, baseURL string, payload map[string]any) diag.Diagnostics {
	var out any
	return client.doJSON(ctx, "POST", joinURL(baseURL, "/instances"), payload, &out)
}

func updateInstance(ctx context.Context, client *APIClient, baseURL, resourceID string, payload map[string]any) diag.Diagnostics {
	return client.doJSON(ctx, "PUT", joinURL(baseURL, fmt.Sprintf("/instances/%s", url.PathEscape(resourceID))), payload, nil)
}

func deleteInstance(ctx context.Context, client *APIClient, baseURL, resourceID string) diag.Diagnostics {
	var diags diag.Diagnostics

	instanceURL := joinURL(baseURL, fmt.Sprintf("/instances/%s", url.PathEscape(resourceID)))
	found, _, delDiags := client.doJSONInternal(ctx, "DELETE", instanceURL, nil, nil, true)
	_ = found
	diags.Append(delDiags...)
	if diags.HasError() {
		return diags
	}

	// Deletions can be eventually consistent. Wait until the instance is not found before
	// returning so dependent deletes (e.g. configurations) don't race.
	const maxAttempts = 8
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var out any
		stillThere, getDiags := client.doJSONAllowNotFound(ctx, "GET", instanceURL, nil, &out)
		diags.Append(getDiags...)
		if diags.HasError() {
			return diags
		}
		if !stillThere {
			return diags
		}

		sleep, ok := retrySleep(ctx, attempt, "", instanceNowFn())
		if !ok {
			return diags
		}
		instanceSleepFn(sleep)
	}

	diags.AddError("Timed out waiting for deletion", "Instance deletion was accepted but the instance is still readable after multiple attempts")
	return diags
}

func readInstanceIntoState(ctx context.Context, client *APIClient, baseURL, resourceID string) (bool, instanceReadResponse, string, diag.Diagnostics) {
	var diags diag.Diagnostics

	var out instanceReadResponse
	found, getDiags := client.doJSONAllowNotFound(
		ctx,
		"GET",
		joinURL(baseURL, fmt.Sprintf("/instances/%s", url.PathEscape(resourceID))),
		nil,
		&out,
	)
	diags.Append(getDiags...)
	if diags.HasError() {
		return false, instanceReadResponse{}, "", diags
	}
	if !found {
		return false, instanceReadResponse{}, "", diags
	}

	if out.Instance.InlineConfiguration == nil {
		return true, out, "", diags
	}

	b, err := jsonMarshal(out.Instance.InlineConfiguration)
	if err != nil {
		// Do not fail the read; inline_configuration is optional/computed.
		return true, out, "", diags
	}

	return true, out, string(b), diags
}
