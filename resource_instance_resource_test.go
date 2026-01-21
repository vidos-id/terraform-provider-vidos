package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func initResourceState(t *testing.T, s *tfsdk.State) {
	t.Helper()
	s.Schema = instanceSchema()
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)
}

func TestInstanceResource_Configure(t *testing.T) {
	r := &instanceResource{}

	// nil ProviderData should be a no-op
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

func TestInstanceResource_Create_SendsExpectedPayload(t *testing.T) {
	var gotBody string
	var calls int

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			if r.Method != http.MethodPost {
				return httpResponse(500, nil, "unexpected"), nil
			}
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(200, nil, `{}`), nil
		}
		// readIntoState GET
		if r.Method != http.MethodGet {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(200, nil, `{"instance":{"resourceId":"tf-abc","name":"n","configurationResourceId":"cid","inlineConfiguration":null}}`), nil
	}))

	oldRead := cryptoRandRead
	cryptoRandRead = func(b []byte) (int, error) {
		// 12 bytes; hex becomes 24 chars.
		for i := range b {
			b[i] = 0xAB
		}
		return len(b), nil
	}
	t.Cleanup(func() { cryptoRandRead = oldRead })

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}

	plan := instanceModel{
		ResourceID:              types.StringNull(),
		Name:                    types.StringValue("n"),
		ConfigurationResourceID: types.StringValue("cid"),
		InlineConfiguration:     types.StringNull(),
	}
	config := instanceModel{ResourceID: types.StringNull()}

	var resp resource.CreateResponse
	initResourceState(t, &resp.State)
	r.Create(context.Background(), resource.CreateRequest{Config: instanceConfig(t, config), Plan: instancePlan(t, plan)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls < 2 {
		t.Fatalf("expected POST+GET, got %d calls", calls)
	}
	if !strings.Contains(gotBody, `"instanceResourceId":"tf-`) {
		t.Fatalf("expected instanceResourceId in payload, got %s", gotBody)
	}
	if !strings.Contains(gotBody, `"name":"n"`) {
		t.Fatalf("expected name in payload, got %s", gotBody)
	}
	if !strings.Contains(gotBody, `"configurationResourceId":"cid"`) {
		t.Fatalf("expected configurationResourceId in payload, got %s", gotBody)
	}
	if strings.Contains(gotBody, "inlineConfiguration") {
		t.Fatalf("expected inlineConfiguration omitted when null")
	}
}

func TestInstanceResource_Create_NullConfigurationResourceIdSendsExplicitNull(t *testing.T) {
	var gotBody string
	var calls int

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(200, nil, `{}`), nil
		}
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"","inlineConfiguration":null}}`), nil
	}))

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}

	plan := instanceModel{
		ResourceID:              types.StringValue("rid"),
		Name:                    types.StringValue("n"),
		ConfigurationResourceID: types.StringNull(),
		InlineConfiguration:     types.StringNull(),
	}
	config := instanceModel{ResourceID: types.StringValue("rid")}

	var resp resource.CreateResponse
	initResourceState(t, &resp.State)
	r.Create(context.Background(), resource.CreateRequest{Config: instanceConfig(t, config), Plan: instancePlan(t, plan)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !strings.Contains(gotBody, `"configurationResourceId":null`) {
		t.Fatalf("expected configurationResourceId:null in payload, got %s", gotBody)
	}
}

func TestInstanceResource_Create_InlineConfigurationInvalidJSONAddsDiagnostics(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("http should not be called on inline JSON parse error")
		return nil, nil
	}))

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}

	plan := instanceModel{
		ResourceID:              types.StringValue("rid"),
		Name:                    types.StringValue("n"),
		ConfigurationResourceID: types.StringNull(),
		InlineConfiguration:     types.StringValue("not-json"),
	}
	config := instanceModel{ResourceID: types.StringValue("rid")}

	var resp resource.CreateResponse
	initResourceState(t, &resp.State)
	r.Create(context.Background(), resource.CreateRequest{Config: instanceConfig(t, config), Plan: instancePlan(t, plan)}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}

func TestInstanceResource_Create_InlineConfigurationIncludedWhenSet(t *testing.T) {
	var gotBody string
	var calls int

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(200, nil, `{}`), nil
		}
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"","inlineConfiguration":{}}}`), nil
	}))

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}
	plan := instanceModel{ResourceID: types.StringValue("rid"), Name: types.StringValue("n"), ConfigurationResourceID: types.StringNull(), InlineConfiguration: types.StringValue(`{"a":1}`)}
	config := instanceModel{ResourceID: types.StringValue("rid")}

	var resp resource.CreateResponse
	initResourceState(t, &resp.State)
	r.Create(context.Background(), resource.CreateRequest{Config: instanceConfig(t, config), Plan: instancePlan(t, plan)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !strings.Contains(gotBody, `"inlineConfiguration":{"a":1}`) {
		t.Fatalf("expected inlineConfiguration object, got %s", gotBody)
	}
}

func TestInstanceResource_Update_OmitsInlineConfigurationWhenNull(t *testing.T) {
	var gotBody string
	var calls int

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ``), nil
		}
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"","inlineConfiguration":null}}`), nil
	}))

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}
	plan := instanceModel{ResourceID: types.StringValue("rid"), Name: types.StringValue("n"), ConfigurationResourceID: types.StringNull(), InlineConfiguration: types.StringNull()}

	var resp resource.UpdateResponse
	initResourceState(t, &resp.State)
	r.Update(context.Background(), resource.UpdateRequest{Plan: instancePlan(t, plan)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls < 2 {
		t.Fatalf("expected PUT+GET, got %d calls", calls)
	}
	if strings.Contains(gotBody, "inlineConfiguration") {
		t.Fatalf("expected inlineConfiguration omitted when null")
	}
}

func TestInstanceResource_Update_OmitsFieldsWhenUnknown(t *testing.T) {
	var gotBody string
	var calls int

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ``), nil
		}
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"","inlineConfiguration":null}}`), nil
	}))

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}
	plan := instanceModel{ResourceID: types.StringValue("rid"), Name: types.StringValue("n"), ConfigurationResourceID: types.StringUnknown(), InlineConfiguration: types.StringUnknown()}

	var resp resource.UpdateResponse
	initResourceState(t, &resp.State)
	r.Update(context.Background(), resource.UpdateRequest{Plan: instancePlan(t, plan)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls < 2 {
		t.Fatalf("expected PUT+GET, got %d calls", calls)
	}
	if strings.Contains(gotBody, "configurationResourceId") {
		t.Fatalf("expected configurationResourceId omitted when unknown")
	}
	if strings.Contains(gotBody, "inlineConfiguration") {
		t.Fatalf("expected inlineConfiguration omitted when unknown")
	}
}

func TestInstanceResource_Update_InvalidInlineConfigurationJSONAddsDiagnostics_NoHTTP(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}
	plan := instanceModel{
		ResourceID:              types.StringValue("rid"),
		Name:                    types.StringValue("n"),
		ConfigurationResourceID: types.StringValue("cid"),
		InlineConfiguration:     types.StringValue(`{bad json}`),
	}

	var resp resource.UpdateResponse
	initResourceState(t, &resp.State)
	r.Update(context.Background(), resource.UpdateRequest{Plan: instancePlan(t, plan)}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestInstanceResource_Update_InlineConfigurationIncludedWhenSet(t *testing.T) {
	var gotBody string
	var calls int

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ``), nil
		}
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"","inlineConfiguration":{}}}`), nil
	}))

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}
	plan := instanceModel{
		ResourceID:              types.StringValue("rid"),
		Name:                    types.StringValue("n"),
		ConfigurationResourceID: types.StringNull(),
		InlineConfiguration:     types.StringValue(`{"a":1}`),
	}

	var resp resource.UpdateResponse
	initResourceState(t, &resp.State)
	r.Update(context.Background(), resource.UpdateRequest{Plan: instancePlan(t, plan)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls < 2 {
		t.Fatalf("expected PUT+GET, got %d calls", calls)
	}
	if !strings.Contains(gotBody, `"inlineConfiguration":{"a":1}`) {
		t.Fatalf("expected inlineConfiguration object, got %s", gotBody)
	}
}

func TestInstanceResource_Update_NullConfigurationResourceIdSendsExplicitNull(t *testing.T) {
	var gotBody string
	var calls int

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ``), nil
		}
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"","inlineConfiguration":null}}`), nil
	}))

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}
	plan := instanceModel{
		ResourceID:              types.StringValue("rid"),
		Name:                    types.StringValue("n"),
		ConfigurationResourceID: types.StringNull(),
		InlineConfiguration:     types.StringNull(),
	}

	var resp resource.UpdateResponse
	initResourceState(t, &resp.State)
	r.Update(context.Background(), resource.UpdateRequest{Plan: instancePlan(t, plan)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls < 2 {
		t.Fatalf("expected PUT+GET, got %d calls", calls)
	}
	if !strings.Contains(gotBody, `"configurationResourceId":null`) {
		t.Fatalf("expected configurationResourceId:null in payload, got %s", gotBody)
	}
}

func TestInstanceResource_Read_NotFoundRemovesResource(t *testing.T) {
	r := &instanceResource{baseURL: func(*APIClient) string { return "https://example.com" }}
	r.client = newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	state := instanceModel{ResourceID: types.StringValue("rid"), Name: types.StringValue("n")}
	var resp resource.ReadResponse
	initResourceState(t, &resp.State)
	r.Read(context.Background(), resource.ReadRequest{State: instanceState(t, state)}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestInstanceResource_Delete_DelegatesToDeleteInstance(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == http.MethodDelete {
			return httpResponse(204, nil, ``), nil
		}
		// return not found immediately after delete to short-circuit wait loop
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	oldSleep := instanceSleepFn
	instanceSleepFn = func(_ time.Duration) {}
	t.Cleanup(func() { instanceSleepFn = oldSleep })

	r := &instanceResource{client: c, baseURL: func(*APIClient) string { return "https://example.com" }}
	state := instanceModel{ResourceID: types.StringValue("rid"), Name: types.StringValue("n")}

	var resp resource.DeleteResponse
	initResourceState(t, &resp.State)
	r.Delete(context.Background(), resource.DeleteRequest{State: instanceState(t, state)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls < 2 {
		t.Fatalf("expected DELETE+GET, got %d calls", calls)
	}
}

func TestInstanceResource_ImportState_Passthrough(t *testing.T) {
	r := &instanceResource{}
	var resp resource.ImportStateResponse
	initResourceState(t, &resp.State)
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "rid"}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	// ImportStatePassthroughID uses State.SetAttribute(), so assert via GetAttribute.
	var got string
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), path.Root("resource_id"), &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if got != "rid" {
		t.Fatalf("unexpected imported id: %q", got)
	}
	// sanity: state should be initialized by SetAttribute
	if resp.State == (tfsdk.State{}) {
		t.Fatalf("expected state to be initialized")
	}
}
