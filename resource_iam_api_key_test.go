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

func initIamApiKeyState(t *testing.T, s *tfsdk.State) {
	t.Helper()
	s.Schema = iamApiKeySchema()
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)
}

func TestIamApiKeyResource_Configure(t *testing.T) {
	r := &IamApiKeyResource{}

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

func TestIamApiKeyResource_Create_InlinePolicyIncludedWhenSet(t *testing.T) {
	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			return httpResponse(500, nil, "unexpected"), nil
		}
		if got := r.URL.String(); got != "https://iam.management.global.example.com/api-keys" {
			return httpResponse(500, nil, "unexpected url: "+got), nil
		}
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		return httpResponse(200, nil, `{"apiKey":{"resourceId":"rid","name":"n","inlinePolicyDocument":{"a":1},"apiSecret":"secret"}}`), nil
	}))

	r := &IamApiKeyResource{client: c}
	planModel := iamApiKeyModel{
		ResourceID:           types.StringUnknown(),
		Name:                 types.StringValue("n"),
		InlinePolicyDocument: types.StringValue(`{"a":1}`),
		ApiSecret:            types.StringUnknown(),
	}

	var req resource.CreateRequest
	req.Plan = iamApiKeyPlan(t, planModel)

	var resp resource.CreateResponse
	initIamApiKeyState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !strings.Contains(gotBody, `"inlinePolicyDocument":{"a":1}`) {
		t.Fatalf("expected inlinePolicyDocument in payload, got: %s", gotBody)
	}
}

func TestIamApiKeyResource_Create_InvalidInlinePolicyJSONAddsDiagnostics(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &IamApiKeyResource{client: c}
	planModel := iamApiKeyModel{
		Name:                 types.StringValue("n"),
		InlinePolicyDocument: types.StringValue(`{bad json}`),
	}

	var req resource.CreateRequest
	req.Plan = iamApiKeyPlan(t, planModel)

	var resp resource.CreateResponse
	initIamApiKeyState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestIamApiKeyResource_Read_PreservesApiSecret(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(200, nil, `{"apiKey":{"resourceId":"rid","name":"n","inlinePolicyDocument":null}}`), nil
	}))

	r := &IamApiKeyResource{client: c}
	stateModel := iamApiKeyModel{
		ResourceID: types.StringValue("rid"),
		ApiSecret:  types.StringValue("keepme"),
		Name:       types.StringValue("old"),
	}

	var req resource.ReadRequest
	req.State = iamApiKeyState(t, stateModel)

	var resp resource.ReadResponse
	initIamApiKeyState(t, &resp.State)

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	var got string
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), path.Root("api_secret"), &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if got != "keepme" {
		t.Fatalf("expected api_secret preserved, got %q", got)
	}
}

func TestIamApiKeyResource_Read_NotFoundRemovesResource(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	r := &IamApiKeyResource{client: c}
	stateModel := iamApiKeyModel{ResourceID: types.StringValue("rid"), ApiSecret: types.StringValue("keepme")}

	var req resource.ReadRequest
	req.State = iamApiKeyState(t, stateModel)

	var resp resource.ReadResponse
	initIamApiKeyState(t, &resp.State)

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestIamApiKeyResource_ReadIntoState_InlinePolicyDocumentPresent(t *testing.T) {
	r := &IamApiKeyResource{}
	r.client = newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(200, nil, `{"apiKey":{"resourceId":"rid","name":"n","inlinePolicyDocument":{"a":1}}}`), nil
	}))

	var state iamApiKeyModel
	found, diags := r.readIntoState(context.Background(), "rid", &state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !found {
		t.Fatalf("expected found")
	}
	if got := state.InlinePolicyDocument.ValueString(); strings.TrimSpace(got) != `{"a":1}` {
		t.Fatalf("unexpected inline_policy_document: %q", got)
	}
}

func TestIamApiKeyResource_ReadIntoState_NotFound(t *testing.T) {
	r := &IamApiKeyResource{}
	r.client = newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	var state iamApiKeyModel
	found, diags := r.readIntoState(context.Background(), "rid", &state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if found {
		t.Fatalf("expected not found")
	}
}

func TestIamApiKeyResource_Update_InlinePolicyNullSendsExplicitNull(t *testing.T) {
	var gotBody string
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPost:
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ""), nil
		case r.Method == http.MethodGet:
			return httpResponse(200, nil, `{"apiKey":{"resourceId":"rid","name":"n2","inlinePolicyDocument":null}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamApiKeyResource{client: c}
	planModel := iamApiKeyModel{
		ResourceID:           types.StringValue("rid"),
		Name:                 types.StringValue("n2"),
		InlinePolicyDocument: types.StringNull(),
		ApiSecret:            types.StringValue("keepme"),
	}

	var req resource.UpdateRequest
	req.Plan = iamApiKeyPlan(t, planModel)

	var resp resource.UpdateResponse
	initIamApiKeyState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected POST+GET, got %d calls", calls)
	}
	if !strings.Contains(gotBody, `"inlinePolicyDocument":null`) {
		t.Fatalf("expected inlinePolicyDocument null in payload, got: %s", gotBody)
	}
}

func TestIamApiKeyResource_Update_InlinePolicyUnknownOmitsInlinePolicyDocument(t *testing.T) {
	var gotBody string
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPost:
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ""), nil
		case r.Method == http.MethodGet:
			return httpResponse(200, nil, `{"apiKey":{"resourceId":"rid","name":"n2","inlinePolicyDocument":null}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamApiKeyResource{client: c}
	planModel := iamApiKeyModel{
		ResourceID:           types.StringValue("rid"),
		Name:                 types.StringValue("n2"),
		InlinePolicyDocument: types.StringUnknown(),
		ApiSecret:            types.StringValue("keepme"),
	}

	var req resource.UpdateRequest
	req.Plan = iamApiKeyPlan(t, planModel)

	var resp resource.UpdateResponse
	initIamApiKeyState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected POST+GET, got %d calls", calls)
	}
	if strings.Contains(gotBody, `"inlinePolicyDocument"`) {
		t.Fatalf("expected inlinePolicyDocument omitted when unknown, got: %s", gotBody)
	}

	var got string
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), path.Root("api_secret"), &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if got != "keepme" {
		t.Fatalf("expected api_secret preserved, got %q", got)
	}
}

func TestIamApiKeyResource_Update_UnknownResourceIDAddsDiagnostics(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &IamApiKeyResource{client: c}
	planModel := iamApiKeyModel{
		ResourceID: types.StringUnknown(),
		Name:       types.StringValue("n"),
	}

	var req resource.UpdateRequest
	req.Plan = iamApiKeyPlan(t, planModel)

	var resp resource.UpdateResponse
	initIamApiKeyState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestIamApiKeyResource_Update_InvalidInlinePolicyJSONAddsDiagnostics_NoHTTP(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &IamApiKeyResource{client: c}
	planModel := iamApiKeyModel{
		ResourceID:           types.StringValue("rid"),
		Name:                 types.StringValue("n"),
		InlinePolicyDocument: types.StringValue(`{bad json}`),
		ApiSecret:            types.StringValue("keepme"),
	}

	var req resource.UpdateRequest
	req.Plan = iamApiKeyPlan(t, planModel)

	var resp resource.UpdateResponse
	initIamApiKeyState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestIamApiKeyResource_Update_ReadBackNotFound_RemovesState(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch r.Method {
		case http.MethodPost:
			return httpResponse(204, nil, ""), nil
		case http.MethodGet:
			return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamApiKeyResource{client: c}
	planModel := iamApiKeyModel{
		ResourceID: types.StringValue("rid"),
		Name:       types.StringValue("n"),
		ApiSecret:  types.StringValue("keepme"),
	}

	var req resource.UpdateRequest
	req.Plan = iamApiKeyPlan(t, planModel)

	var resp resource.UpdateResponse
	initIamApiKeyState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected POST+GET, got %d calls", calls)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestIamApiKeyResource_Delete_AllowNotFoundIsNotAnError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	r := &IamApiKeyResource{client: c}
	stateModel := iamApiKeyModel{ResourceID: types.StringValue("rid")}

	var req resource.DeleteRequest
	req.State = iamApiKeyState(t, stateModel)

	var resp resource.DeleteResponse
	r.Delete(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
}

func TestIamApiKeyResource_ImportState_Passthrough(t *testing.T) {
	r := &IamApiKeyResource{}
	var resp resource.ImportStateResponse
	initIamApiKeyState(t, &resp.State)

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
