package main

import (
	"context"
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

func initIamPolicyState(t *testing.T, s *tfsdk.State) {
	t.Helper()
	s.Schema = iamPolicySchema()
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)
}

func TestIamPolicyResource_Configure(t *testing.T) {
	r := &IamPolicyResource{}

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

func TestIamPolicyResource_Create_SendsExpectedPayload(t *testing.T) {
	var calls int
	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPost:
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(200, nil, `{}`), nil
		case r.Method == http.MethodGet:
			return httpResponse(200, nil, `{"policy":{"resourceId":"rid","name":"n","document":{"a":1},"policyType":"account"}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamPolicyResource{client: c}
	planModel := iamPolicyModel{
		ResourceID: types.StringNull(),
		Name:       types.StringValue("n"),
		Document:   types.StringValue(`{"a":1}`),
	}
	configModel := iamPolicyModel{
		ResourceID: types.StringValue("rid"),
		Name:       types.StringValue("n"),
		Document:   types.StringValue(`{"a":1}`),
	}

	var req resource.CreateRequest
	req.Plan = iamPolicyPlan(t, planModel)
	req.Config = iamPolicyConfig(t, configModel)

	var resp resource.CreateResponse
	initIamPolicyState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected POST+GET, got %d calls", calls)
	}
	if !strings.Contains(gotBody, `"policyResourceId":"rid"`) {
		t.Fatalf("expected policyResourceId in payload, got: %s", gotBody)
	}
}

func TestIamPolicyResource_Create_ReadBackNotFoundAddsDiagnostics(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == http.MethodPost {
			return httpResponse(200, nil, `{}`), nil
		}
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	r := &IamPolicyResource{client: c}
	planModel := iamPolicyModel{Name: types.StringValue("n"), Document: types.StringValue(`{"a":1}`)}
	configModel := iamPolicyModel{ResourceID: types.StringValue("rid")}

	var req resource.CreateRequest
	req.Plan = iamPolicyPlan(t, planModel)
	req.Config = iamPolicyConfig(t, configModel)

	var resp resource.CreateResponse
	initIamPolicyState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 2 {
		t.Fatalf("expected POST+GET, got %d calls", calls)
	}
}

func TestIamPolicyResource_Read_NotFoundRemovesResource(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	r := &IamPolicyResource{client: c}
	stateModel := iamPolicyModel{ResourceID: types.StringValue("rid")}

	var req resource.ReadRequest
	req.State = iamPolicyState(t, stateModel)

	var resp resource.ReadResponse
	initIamPolicyState(t, &resp.State)

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestIamPolicyResource_Update_SendsExpectedPayload(t *testing.T) {
	var calls int
	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPut:
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ""), nil
		case r.Method == http.MethodGet:
			return httpResponse(200, nil, `{"policy":{"resourceId":"rid","name":"n2","document":{"b":2},"policyType":"account"}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamPolicyResource{client: c}
	planModel := iamPolicyModel{
		ResourceID: types.StringValue("rid"),
		Name:       types.StringValue("n2"),
		Document:   types.StringValue(`{"b":2}`),
	}

	var req resource.UpdateRequest
	req.Plan = iamPolicyPlan(t, planModel)

	var resp resource.UpdateResponse
	initIamPolicyState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected PUT+GET, got %d calls", calls)
	}
	if !strings.Contains(gotBody, `"name":"n2"`) {
		t.Fatalf("expected name in payload, got: %s", gotBody)
	}
}

func TestIamPolicyResource_Update_InvalidDocumentJSONAddsDiagnostics_NoHTTP(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &IamPolicyResource{client: c}
	planModel := iamPolicyModel{
		ResourceID: types.StringValue("rid"),
		Name:       types.StringValue("n"),
		Document:   types.StringValue(`{bad json}`),
	}

	var req resource.UpdateRequest
	req.Plan = iamPolicyPlan(t, planModel)

	var resp resource.UpdateResponse
	initIamPolicyState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestIamPolicyResource_Update_ReadBackNotFound_RemovesState(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch r.Method {
		case http.MethodPut:
			return httpResponse(204, nil, ""), nil
		case http.MethodGet:
			return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamPolicyResource{client: c}
	planModel := iamPolicyModel{
		ResourceID: types.StringValue("rid"),
		Name:       types.StringValue("n"),
		Document:   types.StringValue(`{"a":1}`),
	}

	var req resource.UpdateRequest
	req.Plan = iamPolicyPlan(t, planModel)

	var resp resource.UpdateResponse
	initIamPolicyState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected PUT+GET, got %d calls", calls)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestIamPolicyResource_Delete_AllowNotFoundIsNotAnError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	r := &IamPolicyResource{client: c}
	stateModel := iamPolicyModel{ResourceID: types.StringValue("rid")}

	var req resource.DeleteRequest
	req.State = iamPolicyState(t, stateModel)

	var resp resource.DeleteResponse
	r.Delete(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
}

func TestIamPolicyResource_ImportState_Passthrough(t *testing.T) {
	r := &IamPolicyResource{}
	var resp resource.ImportStateResponse
	initIamPolicyState(t, &resp.State)

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
