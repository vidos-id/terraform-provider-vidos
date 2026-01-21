package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func initConfigurationState(t *testing.T, s *tfsdk.State) {
	t.Helper()
	s.Schema = configurationSchema()
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)
}

func TestGatewayConfigurationResource_Configure(t *testing.T) {
	r := &GatewayConfigurationResource{}

	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: nil}, &resource.ConfigureResponse{})
	if r.client != nil {
		t.Fatalf("expected nil client")
	}

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(200, nil, `{}`), nil
	}))
	r.Configure(context.Background(), resource.ConfigureRequest{ProviderData: c}, &resource.ConfigureResponse{})
	if r.client == nil {
		t.Fatalf("expected client set")
	}
}

func TestGatewayConfigurationResource_Create_SendsExpectedPayload(t *testing.T) {
	var calls int
	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPost:
			if got := r.URL.String(); got != "https://gateway.management.eu.example.com/configurations" {
				return httpResponse(500, nil, "unexpected url: "+got), nil
			}
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(200, nil, `{}`), nil
		case r.Method == http.MethodGet:
			return httpResponse(200, nil, `{"configuration":{"resourceId":"rid","name":"n","values":{"a":1}}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &GatewayConfigurationResource{client: c}

	planModel := configurationModel{
		ResourceID: types.StringNull(),
		Name:       types.StringValue("n"),
		Values:     types.StringValue(`{"a":1}`),
	}
	configModel := configurationModel{
		ResourceID: types.StringNull(),
		Name:       types.StringValue("n"),
		Values:     types.StringValue(`{"a":1}`),
	}

	var req resource.CreateRequest
	req.Plan = configurationPlan(t, planModel)
	req.Config = configurationConfig(t, configModel)

	var resp resource.CreateResponse
	initConfigurationState(t, &resp.State)

	oldRead := cryptoRandRead
	cryptoRandRead = func(b []byte) (int, error) {
		copy(b, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
		return len(b), nil
	}
	t.Cleanup(func() { cryptoRandRead = oldRead })

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected POST+GET, got %d calls", calls)
	}
	if !strings.Contains(gotBody, `"configurationResourceId":"tf-000102030405060708090a0b"`) {
		t.Fatalf("expected generated resource id in payload, got: %s", gotBody)
	}
	if !strings.Contains(gotBody, `"name":"n"`) {
		t.Fatalf("expected name in payload, got: %s", gotBody)
	}
	if !strings.Contains(gotBody, `"values":{"a":1}`) {
		t.Fatalf("expected values object in payload, got: %s", gotBody)
	}
}

func TestGatewayConfigurationResource_Create_InvalidValuesJSONAddsDiagnostics(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &GatewayConfigurationResource{client: c}

	planModel := configurationModel{
		ResourceID: types.StringNull(),
		Name:       types.StringValue("n"),
		Values:     types.StringValue(`{bad json}`),
	}
	configModel := configurationModel{
		ResourceID: types.StringNull(),
		Name:       types.StringValue("n"),
		Values:     types.StringValue(`{bad json}`),
	}

	var req resource.CreateRequest
	req.Plan = configurationPlan(t, planModel)
	req.Config = configurationConfig(t, configModel)

	var resp resource.CreateResponse
	initConfigurationState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestGatewayConfigurationResource_Read_NotFoundRemovesResource(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	r := &GatewayConfigurationResource{client: c}
	stateModel := configurationModel{
		ResourceID: types.StringValue("rid"),
		Name:       types.StringValue("n"),
		Values:     types.StringValue(`{"a":1}`),
	}

	var req resource.ReadRequest
	req.State = configurationState(t, stateModel)

	var resp resource.ReadResponse
	initConfigurationState(t, &resp.State)

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestGatewayConfigurationResource_Update_SendsExpectedPayload(t *testing.T) {
	var calls int
	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPut:
			if got := r.URL.String(); got != "https://gateway.management.eu.example.com/configurations/rid" {
				return httpResponse(500, nil, "unexpected url: "+got), nil
			}
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ""), nil
		case r.Method == http.MethodGet:
			return httpResponse(200, nil, `{"configuration":{"resourceId":"rid","name":"n2","values":{"b":2}}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &GatewayConfigurationResource{client: c}

	planModel := configurationModel{
		ResourceID: types.StringValue("rid"),
		Name:       types.StringValue("n2"),
		Values:     types.StringValue(`{"b":2}`),
	}

	var req resource.UpdateRequest
	req.Plan = configurationPlan(t, planModel)

	var resp resource.UpdateResponse
	initConfigurationState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected PUT+GET, got %d calls", calls)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(gotBody), &payload); err != nil {
		t.Fatalf("invalid json payload: %s", err)
	}
	if _, ok := payload["configuration"]; !ok {
		t.Fatalf("expected configuration key in payload: %s", gotBody)
	}
}

func TestGatewayConfigurationResource_Delete_CallsDeleteConfiguration(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != http.MethodDelete {
			return httpResponse(500, nil, "unexpected"), nil
		}
		if got := r.URL.String(); got != "https://gateway.management.eu.example.com/configurations/rid" {
			return httpResponse(500, nil, "unexpected url: "+got), nil
		}
		return httpResponse(204, nil, ""), nil
	}))

	r := &GatewayConfigurationResource{client: c}
	stateModel := configurationModel{ResourceID: types.StringValue("rid")}

	var req resource.DeleteRequest
	req.State = configurationState(t, stateModel)

	var resp resource.DeleteResponse
	r.Delete(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestGatewayConfigurationResource_ImportState_Passthrough(t *testing.T) {
	r := &GatewayConfigurationResource{}
	var resp resource.ImportStateResponse
	initConfigurationState(t, &resp.State)

	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "rid"}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}

	var got string
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), path.Root("resource_id"), &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if got != "rid" {
		t.Fatalf("unexpected imported id: %q", got)
	}
}
