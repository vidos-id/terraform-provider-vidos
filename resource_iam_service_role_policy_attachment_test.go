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

func initIamServiceRolePolicyAttachmentState(t *testing.T, s *tfsdk.State) {
	t.Helper()
	s.Schema = iamServiceRolePolicyAttachmentSchema()
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)
}

func TestIamServiceRolePolicyAttachmentResource_Configure(t *testing.T) {
	r := &IamServiceRolePolicyAttachmentResource{}

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

func TestIamServiceRolePolicyAttachmentResource_Create_InvalidPolicyType(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &IamServiceRolePolicyAttachmentResource{client: c}
	planModel := iamServiceRolePolicyAttachmentModel{
		ServiceRoleID: types.StringValue("sr"),
		PolicyType:    types.StringValue("nope"),
		PolicyID:      types.StringValue("pid"),
	}

	var req resource.CreateRequest
	req.Plan = iamServiceRolePolicyAttachmentPlan(t, planModel)

	var resp resource.CreateResponse
	initIamServiceRolePolicyAttachmentState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestIamServiceRolePolicyAttachmentResource_Create_Put405FallsBackToReplaceAdd(t *testing.T) {
	var calls int
	var gotPostBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/policies/"):
			return httpResponse(200, nil, `{}`), nil
		case r.Method == http.MethodPut:
			return httpResponse(405, nil, `{"code":"MethodNotAllowed","message":"nope"}`), nil
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/service-roles/"):
			// listServiceRolePoliciesFromServiceRole uses includePolicies=true.
			return httpResponse(200, nil, `{"serviceRole":{"policies":[{"policyType":"account","resourceId":"p1"}]}}`), nil
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/policies"):
			b, _ := io.ReadAll(r.Body)
			gotPostBody = string(b)
			var payload map[string]any
			if err := json.Unmarshal(b, &payload); err != nil {
				return httpResponse(400, nil, "bad json"), nil
			}
			return httpResponse(204, nil, ""), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamServiceRolePolicyAttachmentResource{client: c}
	planModel := iamServiceRolePolicyAttachmentModel{
		ServiceRoleID: types.StringValue("sr"),
		PolicyType:    types.StringValue("account"),
		PolicyID:      types.StringValue("pid"),
	}

	var req resource.CreateRequest
	req.Plan = iamServiceRolePolicyAttachmentPlan(t, planModel)

	var resp resource.CreateResponse
	initIamServiceRolePolicyAttachmentState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 4 {
		t.Fatalf("expected policy GET + PUT + list GET + replace POST, got %d calls", calls)
	}
	if !strings.Contains(gotPostBody, `"serviceRolePolicies"`) || !strings.Contains(gotPostBody, `"policyResourceId":"pid"`) {
		t.Fatalf("expected replacement POST body to include pid, got: %s", gotPostBody)
	}
}

func TestIamServiceRolePolicyAttachmentResource_Read_NotAttachedRemovesResource(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// listServiceRolePoliciesFromServiceRole uses GET /service-roles/{id}?includePolicies=true
		return httpResponse(200, nil, `{"serviceRole":{"policies":[]}}`), nil
	}))

	r := &IamServiceRolePolicyAttachmentResource{client: c}
	stateModel := iamServiceRolePolicyAttachmentModel{
		ID:            types.StringValue("sr:account:pid"),
		ServiceRoleID: types.StringValue("sr"),
		PolicyType:    types.StringValue("account"),
		PolicyID:      types.StringValue("pid"),
	}

	var req resource.ReadRequest
	req.State = iamServiceRolePolicyAttachmentState(t, stateModel)

	var resp resource.ReadResponse
	initIamServiceRolePolicyAttachmentState(t, &resp.State)

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestIamServiceRolePolicyAttachmentResource_Read_AttachedKeepsStateAndSetsID(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// listServiceRolePoliciesFromServiceRole uses GET /service-roles/{id}?includePolicies=true
		return httpResponse(200, nil, `{"serviceRole":{"policies":[{"policyType":"managed","resourceId":"pid"}]}}`), nil
	}))

	r := &IamServiceRolePolicyAttachmentResource{client: c}
	stateModel := iamServiceRolePolicyAttachmentModel{
		ServiceRoleID: types.StringValue("sr"),
		PolicyType:    types.StringValue("MANAGED"),
		PolicyID:      types.StringValue("pid"),
	}

	var req resource.ReadRequest
	req.State = iamServiceRolePolicyAttachmentState(t, stateModel)

	var resp resource.ReadResponse
	initIamServiceRolePolicyAttachmentState(t, &resp.State)

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() {
		t.Fatalf("expected state preserved")
	}

	var got string
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), path.Root("id"), &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if got != "sr:managed:pid" {
		t.Fatalf("unexpected id: %q", got)
	}
}

func TestIamServiceRolePolicyAttachmentResource_Update_Unsupported(t *testing.T) {
	r := &IamServiceRolePolicyAttachmentResource{}
	var resp resource.UpdateResponse
	r.Update(context.Background(), resource.UpdateRequest{}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}

func TestIamServiceRolePolicyAttachmentResource_Delete_405FallsBackToReplaceRemove(t *testing.T) {
	var calls int
	var gotPostBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodDelete:
			return httpResponse(405, nil, `{"code":"MethodNotAllowed","message":"nope"}`), nil
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/service-roles/"):
			return httpResponse(200, nil, `{"serviceRole":{"policies":[{"policyType":"account","resourceId":"pid"},{"policyType":"account","resourceId":"other"}]}}`), nil
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/policies"):
			b, _ := io.ReadAll(r.Body)
			gotPostBody = string(b)
			return httpResponse(204, nil, ""), nil
		default:
			return httpResponse(500, nil, "unexpected"), nil
		}
	}))

	r := &IamServiceRolePolicyAttachmentResource{client: c}
	stateModel := iamServiceRolePolicyAttachmentModel{
		ServiceRoleID: types.StringValue("sr"),
		PolicyType:    types.StringValue("account"),
		PolicyID:      types.StringValue("pid"),
	}

	var req resource.DeleteRequest
	req.State = iamServiceRolePolicyAttachmentState(t, stateModel)

	var resp resource.DeleteResponse
	r.Delete(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 3 {
		t.Fatalf("expected DELETE + list GET + replace POST, got %d calls", calls)
	}
	if strings.Contains(gotPostBody, `"policyResourceId":"pid"`) {
		t.Fatalf("expected pid removed from replacement payload, got: %s", gotPostBody)
	}
}

func TestIamServiceRolePolicyAttachmentResource_ImportState_SetsAttributes(t *testing.T) {
	r := &IamServiceRolePolicyAttachmentResource{}
	var resp resource.ImportStateResponse
	initIamServiceRolePolicyAttachmentState(t, &resp.State)

	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "sr:account:pid"}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}

	var got string
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), path.Root("service_role_id"), &got)...)
	if got != "sr" {
		t.Fatalf("unexpected service_role_id: %q", got)
	}
}

func TestIamServiceRolePolicyAttachmentResource_ImportState_InvalidIDAddsDiagnostics(t *testing.T) {
	r := &IamServiceRolePolicyAttachmentResource{}
	var resp resource.ImportStateResponse
	initIamServiceRolePolicyAttachmentState(t, &resp.State)

	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "too:few"}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}
