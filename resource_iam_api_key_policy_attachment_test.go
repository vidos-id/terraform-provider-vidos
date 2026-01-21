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

func initIamApiKeyPolicyAttachmentState(t *testing.T, s *tfsdk.State) {
	t.Helper()
	s.Schema = iamApiKeyPolicyAttachmentSchema()
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(context.Background()), nil)
}

func TestIamApiKeyPolicyAttachmentResource_Configure(t *testing.T) {
	r := &IamApiKeyPolicyAttachmentResource{}

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

func TestIamApiKeyPolicyAttachmentResource_Create_InvalidPolicyTypeAddsDiagnostics(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(500, nil, "unexpected"), nil
	}))

	r := &IamApiKeyPolicyAttachmentResource{client: c}
	planModel := iamApiKeyPolicyAttachmentModel{
		ApiKeyID:   types.StringValue("ak"),
		PolicyType: types.StringValue("nope"),
		PolicyID:   types.StringValue("pid"),
	}

	var req resource.CreateRequest
	req.Plan = iamApiKeyPolicyAttachmentPlan(t, planModel)

	var resp resource.CreateResponse
	initIamApiKeyPolicyAttachmentState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
	if calls != 0 {
		t.Fatalf("expected no http calls, got %d", calls)
	}
}

func TestIamApiKeyPolicyAttachmentResource_Create_PUTFallback405UsesReplacePoliciesAdd(t *testing.T) {
	var gotReplaceBody string
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/policies/"):
			// getPolicy
			return httpResponse(200, nil, `{}`), nil
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/api-keys/"):
			// attach attempt (not supported)
			return httpResponse(405, nil, `{"code":"MethodNotAllowed"}`), nil
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/policies"):
			// listApiKeyPolicies
			return httpResponse(200, nil, `{"apiKeyPolicies":[]}`), nil
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/policies"):
			b, _ := io.ReadAll(r.Body)
			gotReplaceBody = string(b)
			return httpResponse(204, nil, ""), nil
		default:
			return httpResponse(500, nil, "unexpected "+r.Method+" "+r.URL.String()), nil
		}
	}))

	r := &IamApiKeyPolicyAttachmentResource{client: c}
	planModel := iamApiKeyPolicyAttachmentModel{
		ApiKeyID:   types.StringValue("ak"),
		PolicyType: types.StringValue("account"),
		PolicyID:   types.StringValue("pid"),
	}

	var req resource.CreateRequest
	req.Plan = iamApiKeyPolicyAttachmentPlan(t, planModel)

	var resp resource.CreateResponse
	initIamApiKeyPolicyAttachmentState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 4 {
		// GET policy + PUT attach + GET list + POST replace
		t.Fatalf("expected 4 http calls, got %d", calls)
	}
	if !strings.Contains(gotReplaceBody, `"policyType":"account"`) {
		t.Fatalf("expected policyType in replace payload, got: %s", gotReplaceBody)
	}
	if !strings.Contains(gotReplaceBody, `"policyResourceId":"pid"`) {
		t.Fatalf("expected policyResourceId in replace payload, got: %s", gotReplaceBody)
	}

	var got string
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), path.Root("id"), &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if got != "ak:account:pid" {
		t.Fatalf("unexpected id: %q", got)
	}
}

func TestIamApiKeyPolicyAttachmentResource_Create_PUTFallback405ReplacePoliciesAdd_DoesNotDuplicateExisting(t *testing.T) {
	var gotReplaceBody string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/policies/"):
			return httpResponse(200, nil, `{}`), nil
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/api-keys/"):
			return httpResponse(405, nil, `{"code":"MethodNotAllowed"}`), nil
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/policies"):
			// listApiKeyPolicies already includes the policy.
			return httpResponse(200, nil, `{"apiKeyPolicies":[{"policyType":"account","policyResourceId":"pid"}]}`), nil
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/policies"):
			b, _ := io.ReadAll(r.Body)
			gotReplaceBody = string(b)
			return httpResponse(204, nil, ""), nil
		default:
			return httpResponse(500, nil, "unexpected "+r.Method+" "+r.URL.String()), nil
		}
	}))

	r := &IamApiKeyPolicyAttachmentResource{client: c}
	planModel := iamApiKeyPolicyAttachmentModel{
		ApiKeyID:   types.StringValue("ak"),
		PolicyType: types.StringValue("account"),
		PolicyID:   types.StringValue("pid"),
	}

	var req resource.CreateRequest
	req.Plan = iamApiKeyPolicyAttachmentPlan(t, planModel)

	var resp resource.CreateResponse
	initIamApiKeyPolicyAttachmentState(t, &resp.State)

	r.Create(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}

	if strings.Count(gotReplaceBody, `"policyResourceId":"pid"`) != 1 {
		t.Fatalf("expected pid only once in payload, got: %s", gotReplaceBody)
	}
}

func TestIamApiKeyPolicyAttachmentResource_Read_NotAttachedRemovesResource(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// listApiKeyPolicies
		return httpResponse(200, nil, `{"apiKeyPolicies":[]}`), nil
	}))

	r := &IamApiKeyPolicyAttachmentResource{client: c}
	stateModel := iamApiKeyPolicyAttachmentModel{
		ID:         types.StringValue("ak:account:pid"),
		ApiKeyID:   types.StringValue("ak"),
		PolicyType: types.StringValue("account"),
		PolicyID:   types.StringValue("pid"),
	}

	var req resource.ReadRequest
	req.State = iamApiKeyPolicyAttachmentState(t, stateModel)

	var resp resource.ReadResponse
	initIamApiKeyPolicyAttachmentState(t, &resp.State)

	r.Read(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Fatalf("expected state removed")
	}
}

func TestIamApiKeyPolicyAttachmentResource_Read_AttachedKeepsStateAndSetsID(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// listApiKeyPolicies
		return httpResponse(200, nil, `{"apiKeyPolicies":[{"policyType":"account","policyResourceId":"pid"}]}`), nil
	}))

	r := &IamApiKeyPolicyAttachmentResource{client: c}
	stateModel := iamApiKeyPolicyAttachmentModel{
		ApiKeyID:   types.StringValue("ak"),
		PolicyType: types.StringValue("ACCOUNT"),
		PolicyID:   types.StringValue("pid"),
	}

	var req resource.ReadRequest
	req.State = iamApiKeyPolicyAttachmentState(t, stateModel)

	var resp resource.ReadResponse
	initIamApiKeyPolicyAttachmentState(t, &resp.State)

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
	if got != "ak:account:pid" {
		t.Fatalf("unexpected id: %q", got)
	}
}

func TestIamApiKeyPolicyAttachmentResource_Update_IsUnsupported(t *testing.T) {
	r := &IamApiKeyPolicyAttachmentResource{}
	var resp resource.UpdateResponse
	r.Update(context.Background(), resource.UpdateRequest{}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}

func TestIamApiKeyPolicyAttachmentResource_Delete_PUTFallback405UsesReplacePoliciesRemove(t *testing.T) {
	var gotReplaceBody string
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		switch {
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/api-keys/"):
			// detach attempt (not supported)
			return httpResponse(405, nil, `{"code":"MethodNotAllowed"}`), nil
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/policies"):
			// listApiKeyPolicies returns a policy we will remove
			return httpResponse(200, nil, `{"apiKeyPolicies":[{"policyType":"account","policyResourceId":"pid"},{"policyType":"managed","policyResourceId":"keep"}]}`), nil
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/policies"):
			b, _ := io.ReadAll(r.Body)
			gotReplaceBody = string(b)
			return httpResponse(204, nil, ""), nil
		default:
			return httpResponse(500, nil, "unexpected "+r.Method+" "+r.URL.String()), nil
		}
	}))

	r := &IamApiKeyPolicyAttachmentResource{client: c}
	stateModel := iamApiKeyPolicyAttachmentModel{
		ApiKeyID:   types.StringValue("ak"),
		PolicyType: types.StringValue("account"),
		PolicyID:   types.StringValue("pid"),
	}

	var req resource.DeleteRequest
	req.State = iamApiKeyPolicyAttachmentState(t, stateModel)

	var resp resource.DeleteResponse
	r.Delete(context.Background(), req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if calls != 3 {
		// DELETE detach + GET list + POST replace
		t.Fatalf("expected 3 http calls, got %d", calls)
	}
	if strings.Contains(gotReplaceBody, `"policyResourceId":"pid"`) {
		t.Fatalf("expected removed policy not present, got: %s", gotReplaceBody)
	}
	if !strings.Contains(gotReplaceBody, `"policyResourceId":"keep"`) {
		t.Fatalf("expected remaining policy kept, got: %s", gotReplaceBody)
	}
}

func TestIamApiKeyPolicyAttachmentResource_ImportState_SetsAttributes(t *testing.T) {
	r := &IamApiKeyPolicyAttachmentResource{}
	var resp resource.ImportStateResponse
	initIamApiKeyPolicyAttachmentState(t, &resp.State)

	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "ak:account:pid"}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}

	var got string
	resp.Diagnostics.Append(resp.State.GetAttribute(context.Background(), path.Root("api_key_id"), &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", resp.Diagnostics)
	}
	if got != "ak" {
		t.Fatalf("unexpected api_key_id: %q", got)
	}
}

func TestIamApiKeyPolicyAttachmentResource_ImportState_InvalidIDAddsDiagnostics(t *testing.T) {
	r := &IamApiKeyPolicyAttachmentResource{}
	var resp resource.ImportStateResponse
	initIamApiKeyPolicyAttachmentState(t, &resp.State)

	r.ImportState(context.Background(), resource.ImportStateRequest{ID: "too:few"}, &resp)
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected diagnostics error")
	}
}

func TestComposeAttachmentID_LowercasesPolicyType(t *testing.T) {
	if got := composeAttachmentID("ak", "AcCount", "pid"); got != "ak:account:pid" {
		t.Fatalf("unexpected id: %q", got)
	}
}
