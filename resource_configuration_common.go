package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type configurationModel struct {
	ResourceID types.String `tfsdk:"resource_id"`
	Name       types.String `tfsdk:"name"`
	Values     types.String `tfsdk:"values"`
}

type configurationReadResponse struct {
	Configuration struct {
		ResourceID string `json:"resourceId"`
		Name       string `json:"name"`
		Values     any    `json:"values"`
	} `json:"configuration"`
}

func configurationCreatePayload(resourceID, name string, values any) map[string]any {
	return map[string]any{
		"configurationResourceId": resourceID,
		"configuration": map[string]any{
			"name":   name,
			"values": values,
		},
	}
}

func configurationUpdatePayload(name string, values any) map[string]any {
	return map[string]any{
		"configuration": map[string]any{
			"name":   name,
			"values": values,
		},
	}
}

func createConfiguration(ctx context.Context, client *APIClient, baseURL string, payload map[string]any) diag.Diagnostics {
	var out any
	return client.doJSON(ctx, "POST", joinURL(baseURL, "/configurations"), payload, &out)
}

func updateConfiguration(ctx context.Context, client *APIClient, baseURL, resourceID string, payload map[string]any) diag.Diagnostics {
	return client.doJSON(ctx, "PUT", joinURL(baseURL, fmt.Sprintf("/configurations/%s", url.PathEscape(resourceID))), payload, nil)
}

func deleteConfiguration(ctx context.Context, client *APIClient, baseURL, resourceID string) diag.Diagnostics {
	configurationURL := joinURL(baseURL, fmt.Sprintf("/configurations/%s", url.PathEscape(resourceID)))
	_, status, diags := client.doJSONInternal(ctx, "DELETE", configurationURL, nil, nil, true)
	if diags.HasError() {
		isInUse := false
		for _, d := range diags.Errors() {
			detail := d.Detail()
			// The API sometimes returns "InUse" as HTTP 409, but it has also been observed
			// as a 5xx with a friendly error code. Keep guidance keyed off the code.
			if strings.Contains(detail, "(InUse)") || strings.Contains(detail, "\"code\":\"InUse\"") {
				isInUse = true
				break
			}
		}
		if status == 409 || isInUse {
			diags.AddError(
				"Configuration still in use",
				"This configuration is still referenced by one or more instances. Terraform must first update those instances to stop using it (set configuration_resource_id to null or a different configuration), then delete the configuration in a subsequent apply.",
			)
		}
		return diags
	}

	return diags
}

func readConfigurationIntoState(ctx context.Context, client *APIClient, baseURL, resourceID string) (bool, configurationReadResponse, string, diag.Diagnostics) {
	var diags diag.Diagnostics

	var out configurationReadResponse
	found, getDiags := client.doJSONAllowNotFound(
		ctx,
		"GET",
		joinURL(baseURL, fmt.Sprintf("/configurations/%s", url.PathEscape(resourceID))),
		nil,
		&out,
	)
	diags.Append(getDiags...)
	if diags.HasError() {
		return false, configurationReadResponse{}, "", diags
	}
	if !found {
		return false, configurationReadResponse{}, "", diags
	}

	cfg, err := jsonMarshal(out.Configuration.Values)
	if err != nil {
		diags.AddError("Configuration encode error", err.Error())
		return false, configurationReadResponse{}, "", diags
	}

	return true, out, string(cfg), diags
}
