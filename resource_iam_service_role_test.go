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

func initIamServiceRoleState(t *testing.T, s *tfsdk.State) {
	t.Helper()
	s.Schema = iamServiceRoleSchema()
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)
}

func TestIamServiceRoleResource_Configure(t *testing.T) {
	r := &IamServiceRoleResource{}

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

func TestIamServiceRoleResource_Create_SendsExpectedPayload(t *testing.T) {
	var calls int
	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPost:
			if got := r.URL.String(); got != "https://iam.management.global.example.com/service-roles" {
				return httpResponse(500, nil, "unexpected url: "+got), nil
			}
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(200, nil, `{}`), nil
		case r.Method == http.MethodGet:
			if got := r.URL.String(); got != "https://iam.management.global.example.com/service-roles/rid?resourceOwner=account" {
				return httpResponse(500, nil, "unexpected url: "+got), nil
			}
			return httpResponse(200, nil, `{"serviceRole":{"resourceId":"rid","name":"n","inlinePolicyDocument":{"a":1}}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamServiceRoleResource{client: c}
	configModel := iamServiceRoleModel{
		ResourceID:           types.StringValue("rid"),
		Name:                 types.StringValue("n"),
		InlinePolicyDocument: types.StringNull(),
	}
	planModel := iamServiceRoleModel{
		ResourceID:           types.StringUnknown(),
		Name:                 types.StringValue("n"),
		InlinePolicyDocument: types.StringValue(`{"a":1}`),
	}

	var req resource.CreateRequest
	req.Config = iamServiceRoleConfig(t, configModel)
	req.Plan = iamServiceRolePlan(t, planModel)

	var resp resource.CreateResponse
	initIamServiceRoleState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 2 {
		t.Fatalf("expected POST+GET, got %d calls", calls)
	}
	if !strings.Contains(gotBody, `"serviceRoleResourceId":"rid"`) {
		t.Fatalf("expected serviceRoleResourceId in payload, got: %s", gotBody)
	}
	if !strings.Contains(gotBody, `"inlinePolicyDocument":{"a":1}`) {
		t.Fatalf("expected inlinePolicyDocument object in payload, got: %s", gotBody)
	}
}

func TestIamServiceRoleResource_Create_GeneratesResourceIDWhenConfigNull(t *testing.T) {
	oldRead := cryptoRandRead
	cryptoRandRead = func(b []byte) (int, error) {
		copy(b, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
		return len(b), nil
	}
	t.Cleanup(func() { cryptoRandRead = oldRead })

	expectedID := "tf-000102030405060708090a0b"

	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.Method == http.MethodPost:
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(200, nil, `{}`), nil
		case r.Method == http.MethodGet:
			if got := r.URL.String(); got != "https://iam.management.global.example.com/service-roles/"+expectedID+"?resourceOwner=account" {
				return httpResponse(500, nil, "unexpected url: "+got), nil
			}
			return httpResponse(200, nil, `{"serviceRole":{"resourceId":"`+expectedID+`","name":"n","inlinePolicyDocument":null}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamServiceRoleResource{client: c}
	configModel := iamServiceRoleModel{ResourceID: types.StringNull(), Name: types.StringValue("n")}
	planModel := iamServiceRoleModel{ResourceID: types.StringUnknown(), Name: types.StringValue("n")}

	var req resource.CreateRequest
	req.Config = iamServiceRoleConfig(t, configModel)
	req.Plan = iamServiceRolePlan(t, planModel)

	var resp resource.CreateResponse
	initIamServiceRoleState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !strings.Contains(gotBody, `"serviceRoleResourceId":"`+expectedID+`"`) {
		t.Fatalf("expected generated id in payload, got: %s", gotBody)
	}
}

func TestIamServiceRoleResource_Create_InvalidInlinePolicyJSONAddsDiagnostics(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &IamServiceRoleResource{client: c}
	configModel := iamServiceRoleModel{ResourceID: types.StringValue("rid"), Name: types.StringValue("n")}
	planModel := iamServiceRoleModel{
		ResourceID:           types.StringUnknown(),
		Name:                 types.StringValue("n"),
		InlinePolicyDocument: types.StringValue(`{bad json}`),
	}

	var req resource.CreateRequest
	req.Config = iamServiceRoleConfig(t, configModel)
	req.Plan = iamServiceRolePlan(t, planModel)

	var resp resource.CreateResponse
	initIamServiceRoleState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestIamServiceRoleResource_Read_NotFoundRemovesResource(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	r := &IamServiceRoleResource{client: c}
	stateModel := iamServiceRoleModel{ResourceID: types.StringValue("rid"), Name: types.StringValue("n")}

	var req resource.ReadRequest
	req.State = iamServiceRoleState(t, stateModel)

	var resp resource.ReadResponse
	initIamServiceRoleState(t, &resp.State)

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestIamServiceRoleResource_Update_InlinePolicyNullSendsExplicitNull(t *testing.T) {
	var calls int
	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPost:
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ""), nil
		case r.Method == http.MethodGet:
			return httpResponse(200, nil, `{"serviceRole":{"resourceId":"rid","name":"n2","inlinePolicyDocument":null}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamServiceRoleResource{client: c}
	planModel := iamServiceRoleModel{
		ResourceID:           types.StringValue("rid"),
		Name:                 types.StringValue("n2"),
		InlinePolicyDocument: types.StringNull(),
	}

	var req resource.UpdateRequest
	req.Plan = iamServiceRolePlan(t, planModel)

	var resp resource.UpdateResponse
	initIamServiceRoleState(t, &resp.State)

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

func TestIamServiceRoleResource_Update_InlinePolicyUnknownOmitsInlinePolicyDocument(t *testing.T) {
	var calls int
	var gotBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodPost:
			b, _ := io.ReadAll(r.Body)
			gotBody = string(b)
			return httpResponse(204, nil, ""), nil
		case r.Method == http.MethodGet:
			return httpResponse(200, nil, `{"serviceRole":{"resourceId":"rid","name":"n2","inlinePolicyDocument":null}}`), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamServiceRoleResource{client: c}
	planModel := iamServiceRoleModel{
		ResourceID:           types.StringValue("rid"),
		Name:                 types.StringValue("n2"),
		InlinePolicyDocument: types.StringUnknown(),
	}

	var req resource.UpdateRequest
	req.Plan = iamServiceRolePlan(t, planModel)

	var resp resource.UpdateResponse
	initIamServiceRoleState(t, &resp.State)

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
}

func TestIamServiceRoleResource_Update_InvalidInlinePolicyJSONAddsDiagnostics_NoHTTP(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &IamServiceRoleResource{client: c}
	planModel := iamServiceRoleModel{
		ResourceID:           types.StringValue("rid"),
		Name:                 types.StringValue("n"),
		InlinePolicyDocument: types.StringValue(`{bad json}`),
	}

	var req resource.UpdateRequest
	req.Plan = iamServiceRolePlan(t, planModel)

	var resp resource.UpdateResponse
	initIamServiceRoleState(t, &resp.State)

	r.Update(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestIamServiceRoleResource_Update_ReadBackNotFound_RemovesState(t *testing.T) {
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

	r := &IamServiceRoleResource{client: c}
	planModel := iamServiceRoleModel{
		ResourceID: types.StringValue("rid"),
		Name:       types.StringValue("n"),
	}

	var req resource.UpdateRequest
	req.Plan = iamServiceRolePlan(t, planModel)

	var resp resource.UpdateResponse
	initIamServiceRoleState(t, &resp.State)

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

func TestIamServiceRoleResource_Delete_AllowNotFoundIsNotAnError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodDelete {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	r := &IamServiceRoleResource{client: c}
	stateModel := iamServiceRoleModel{ResourceID: types.StringValue("rid"), Name: types.StringValue("n")}

	var req resource.DeleteRequest
	req.State = iamServiceRoleState(t, stateModel)

	var resp resource.DeleteResponse
	r.Delete(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
}

func TestIamServiceRoleResource_ImportState_Passthrough(t *testing.T) {
	r := &IamServiceRoleResource{}
	var resp resource.ImportStateResponse
	initIamServiceRoleState(t, &resp.State)

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

func TestIamServiceRoleResource_ReadIntoState_Mapping(t *testing.T) {
	r := &IamServiceRoleResource{}
	r.client = newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(200, nil, `{"serviceRole":{"resourceId":"rid","name":"n","inlinePolicyDocument":{"a":1}}}`), nil
	}))

	var state iamServiceRoleModel
	found, diags := r.readIntoState(context.Background(), "rid", &state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !found {
		t.Fatalf("expected found")
	}
	if got := state.ResourceID.ValueString(); got != "rid" {
		t.Fatalf("unexpected resource_id: %q", got)
	}
	if got := state.Name.ValueString(); got != "n" {
		t.Fatalf("unexpected name: %q", got)
	}
	if got := strings.TrimSpace(state.InlinePolicyDocument.ValueString()); got != `{"a":1}` {
		t.Fatalf("unexpected inline_policy_document: %q", got)
	}
}
