package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestConfigurationPayloadBuilders(t *testing.T) {
	p := configurationCreatePayload("rid", "name", map[string]any{"a": 1})
	if p["configurationResourceId"].(string) != "rid" {
		t.Fatalf("unexpected resource id: %#v", p)
	}

	cu, ok := p["configuration"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected configuration type: %T", p["configuration"])
	}
	if cu["name"].(string) != "name" {
		t.Fatalf("unexpected name")
	}

	up := configurationUpdatePayload("n2", map[string]any{"b": 2})
	uu := up["configuration"].(map[string]any)
	if uu["name"].(string) != "n2" {
		t.Fatalf("unexpected update name")
	}
}

func TestInstancePayloadBuilders(t *testing.T) {
	p := instanceCreatePayload("rid", map[string]any{"name": "x"})
	if p["instanceResourceId"].(string) != "rid" {
		t.Fatalf("unexpected resource id: %#v", p)
	}
	if _, ok := p["instance"].(map[string]any); !ok {
		t.Fatalf("unexpected instance type: %T", p["instance"])
	}

	up := instanceUpdatePayload(map[string]any{"name": "y"})
	if _, ok := up["instance"].(map[string]any); !ok {
		t.Fatalf("unexpected instance type: %T", up["instance"])
	}
}

func TestReadConfigurationIntoState_JSONMarshalError(t *testing.T) {
	oldMarshal := jsonMarshal
	jsonMarshal = func(any) ([]byte, error) { return nil, errors.New("nope") }
	t.Cleanup(func() { jsonMarshal = oldMarshal })

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(200, nil, `{"configuration":{"resourceId":"rid","name":"n","values":{"a":1}}}`), nil
	}))

	found, _, _, diags := readConfigurationIntoState(context.Background(), c, "https://example.com", "rid")
	if found {
		t.Fatalf("expected found=false on marshal error")
	}
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestReadConfigurationIntoState_NotFound(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	found, _, _, diags := readConfigurationIntoState(context.Background(), c, "https://example.com", "rid")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if found {
		t.Fatalf("expected found=false")
	}
}

func TestDeleteConfiguration_500ShowsHelpfulMessage(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(500, nil, `{"code":"InUse","message":"still in use"}`), nil
	}))

	diags := deleteConfiguration(context.Background(), c, "https://example.com", "rid")
	if !diags.HasError() {
		t.Fatalf("expected error")
	}
	// First error is from doJSONInternal; we then add a second error with guidance.
	if got := diags.Errors()[0].Summary(); got != "API error" {
		t.Fatalf("unexpected summary: %q", got)
	}
	if len(diags.Errors()) < 2 {
		t.Fatalf("expected guidance diagnostic, got %#v", diags)
	}
	if got := diags.Errors()[1].Summary(); got != "Configuration still in use" {
		t.Fatalf("unexpected guidance summary: %q", got)
	}
}

func TestDeleteConfiguration_409ShowsHelpfulMessage(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(409, nil, `{"code":"InUse","message":"still in use"}`), nil
	}))

	diags := deleteConfiguration(context.Background(), c, "https://example.com", "rid")
	if !diags.HasError() {
		t.Fatalf("expected error")
	}
	if len(diags.Errors()) < 2 {
		t.Fatalf("expected guidance diagnostic, got %#v", diags)
	}
	if got := diags.Errors()[1].Summary(); got != "Configuration still in use" {
		t.Fatalf("unexpected guidance summary: %q", got)
	}
}

func TestDeleteConfiguration_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != "DELETE" {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(204, nil, ""), nil
	}))

	diags := deleteConfiguration(context.Background(), c, "https://example.com", "rid")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
}

func TestReadConfigurationIntoState_Success(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(200, nil, `{"configuration":{"resourceId":"rid","name":"n","values":{"a":1}}}`), nil
	}))

	found, out, valuesJSON, diags := readConfigurationIntoState(context.Background(), c, "https://example.com", "rid")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !found {
		t.Fatalf("expected found")
	}
	if out.Configuration.ResourceID != "rid" || out.Configuration.Name != "n" {
		t.Fatalf("unexpected out: %#v", out)
	}
	if strings.TrimSpace(valuesJSON) != `{"a":1}` {
		t.Fatalf("unexpected values json: %q", valuesJSON)
	}
}

func TestConfigurationRequestHelpers_HitExpectedEndpoints(t *testing.T) {
	var seen []string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		seen = append(seen, r.Method+" "+r.URL.String())
		if r.Method == "POST" {
			b, _ := io.ReadAll(r.Body)
			var v map[string]any
			if err := json.Unmarshal(b, &v); err != nil {
				return httpResponse(400, nil, "bad json"), nil
			}
			if v["configurationResourceId"].(string) != "rid" {
				return httpResponse(400, nil, "bad rid"), nil
			}
			return httpResponse(200, nil, `{}`), nil
		}
		if r.Method == "PUT" {
			return httpResponse(204, nil, ""), nil
		}
		return httpResponse(500, nil, "unexpected"), nil
	}))

	createDiags := createConfiguration(context.Background(), c, "https://example.com", configurationCreatePayload("rid", "n", map[string]any{"a": 1}))
	if createDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", createDiags)
	}

	updateDiags := updateConfiguration(context.Background(), c, "https://example.com", "a b", configurationUpdatePayload("n2", map[string]any{"b": 2}))
	if updateDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", updateDiags)
	}

	if len(seen) != 2 {
		t.Fatalf("expected 2 requests, got %d: %#v", len(seen), seen)
	}
	if seen[0] != "POST https://example.com/configurations" {
		t.Fatalf("unexpected create url: %q", seen[0])
	}
	if seen[1] != "PUT https://example.com/configurations/a%20b" {
		t.Fatalf("unexpected update url: %q", seen[1])
	}
}

func TestInstanceRequestHelpers_HitExpectedEndpoints(t *testing.T) {
	var seen []string
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		seen = append(seen, r.Method+" "+r.URL.String())
		if r.Method == "POST" {
			return httpResponse(200, nil, `{}`), nil
		}
		if r.Method == "PUT" {
			return httpResponse(204, nil, ""), nil
		}
		return httpResponse(500, nil, "unexpected"), nil
	}))

	createDiags := createInstance(context.Background(), c, "https://example.com", instanceCreatePayload("rid", map[string]any{"name": "x"}))
	if createDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", createDiags)
	}

	updateDiags := updateInstance(context.Background(), c, "https://example.com", "a b", instanceUpdatePayload(map[string]any{"name": "y"}))
	if updateDiags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", updateDiags)
	}

	if len(seen) != 2 {
		t.Fatalf("expected 2 requests, got %d: %#v", len(seen), seen)
	}
	if seen[0] != "POST https://example.com/instances" {
		t.Fatalf("unexpected create url: %q", seen[0])
	}
	if seen[1] != "PUT https://example.com/instances/a%20b" {
		t.Fatalf("unexpected update url: %q", seen[1])
	}
}

func TestReadInstanceIntoState_InlineConfigurationMarshalErrorIgnored(t *testing.T) {
	oldMarshal := jsonMarshal
	jsonMarshal = func(any) ([]byte, error) { return nil, errors.New("nope") }
	t.Cleanup(func() { jsonMarshal = oldMarshal })

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"cid","inlineConfiguration":{"a":1}}}`), nil
	}))

	found, _, inline, diags := readInstanceIntoState(context.Background(), c, "https://example.com", "rid")
	if !found || diags.HasError() {
		t.Fatalf("unexpected: found=%v diags=%#v", found, diags)
	}
	if inline != "" {
		t.Fatalf("expected inline_configuration empty string on marshal error")
	}
}

func TestReadInstanceIntoState_NotFound(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	found, _, inline, diags := readInstanceIntoState(context.Background(), c, "https://example.com", "rid")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if found {
		t.Fatalf("expected found=false")
	}
	if inline != "" {
		t.Fatalf("expected inline_configuration empty, got %q", inline)
	}
}

func TestReadInstanceIntoState_DoJSONError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(500, nil, `{"code":"Boom","message":"nope"}`), nil
	}))

	found, _, _, diags := readInstanceIntoState(context.Background(), c, "https://example.com", "rid")
	if found {
		t.Fatalf("expected found=false")
	}
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestReadInstanceIntoState_SuccessInlineConfigurationMarshaled(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"cid","inlineConfiguration":{"a":1}}}`), nil
	}))

	found, out, inline, diags := readInstanceIntoState(context.Background(), c, "https://example.com", "rid")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !found {
		t.Fatalf("expected found")
	}
	if out.Instance.ResourceID != "rid" {
		t.Fatalf("unexpected resource id: %q", out.Instance.ResourceID)
	}
	if strings.TrimSpace(inline) != `{"a":1}` {
		t.Fatalf("unexpected inline_configuration: %q", inline)
	}
}

func TestDeleteInstance_EventuallyConsistentWaitsUntilNotFound(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		// First request is DELETE, then GET until 404.
		calls++
		if r.Method == "DELETE" {
			return httpResponse(204, nil, ""), nil
		}
		if calls < 4 {
			return httpResponse(200, nil, `{}`), nil
		}
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	oldSleep := instanceSleepFn
	oldNow := instanceNowFn
	instanceSleepFn = func(time.Duration) {}
	instanceNowFn = func() time.Time { return time.Unix(0, 0) }
	t.Cleanup(func() {
		instanceSleepFn = oldSleep
		instanceNowFn = oldNow
	})

	diags := deleteInstance(context.Background(), c, "https://example.com", "rid")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if calls < 4 {
		t.Fatalf("expected multiple calls, got %d", calls)
	}
}

func TestDeleteInstance_DeleteErrorStopsEarly(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method != "DELETE" {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(500, nil, `{"code":"Boom","message":"nope"}`), nil
	}))

	diags := deleteInstance(context.Background(), c, "https://example.com", "rid")
	if !diags.HasError() {
		t.Fatalf("expected error")
	}
	if calls != 1 {
		t.Fatalf("expected only DELETE attempt, got %d", calls)
	}
}

func TestDeleteInstance_ContextCancelledStopsWaiting(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == "DELETE" {
			return httpResponse(204, nil, ""), nil
		}
		// Cancel before the retrySleep call so it returns ok=false.
		cancel()
		return httpResponse(200, nil, `{}`), nil
	}))

	oldSleep := instanceSleepFn
	instanceSleepFn = func(time.Duration) { t.Fatalf("sleep should not be called when ctx cancelled") }
	t.Cleanup(func() { instanceSleepFn = oldSleep })

	diags := deleteInstance(ctx, c, "https://example.com", "rid")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if calls != 2 {
		t.Fatalf("expected DELETE + single GET, got %d", calls)
	}
}

func TestDeleteInstance_TimesOutWhenStillReadable(t *testing.T) {
	var getCalls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == "DELETE" {
			return httpResponse(204, nil, ""), nil
		}
		getCalls++
		return httpResponse(200, nil, `{}`), nil
	}))

	oldSleep := instanceSleepFn
	instanceSleepFn = func(time.Duration) {}
	t.Cleanup(func() { instanceSleepFn = oldSleep })

	diags := deleteInstance(context.Background(), c, "https://example.com", "rid")
	if !diags.HasError() {
		t.Fatalf("expected timeout diagnostics")
	}
	if !strings.Contains(diags.Errors()[0].Summary(), "Timed out") {
		// summary is "Timed out waiting for deletion".
		// Keep this loose so minor text changes don't break tests.
		t.Fatalf("unexpected summary: %q", diags.Errors()[0].Summary())
	}
	if getCalls != 8 {
		t.Fatalf("expected 8 GET attempts, got %d", getCalls)
	}
}

func TestInstanceResource_ReadIntoState_Mapping(t *testing.T) {
	r := &instanceResource{baseURL: func(*APIClient) string { return "https://example.com" }}
	r.client = newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != "GET" {
			return httpResponse(500, nil, "unexpected"), nil
		}
		return httpResponse(200, nil, `{"instance":{"resourceId":"rid","name":"n","configurationResourceId":"","inlineConfiguration":null}}`), nil
	}))

	var state instanceModel
	found, diags := r.readIntoState(context.Background(), "rid", &state)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !found {
		t.Fatalf("expected found")
	}
	if state.ResourceID.ValueString() != "rid" {
		t.Fatalf("unexpected resource id: %q", state.ResourceID.ValueString())
	}
	if !state.ConfigurationResourceID.IsNull() {
		t.Fatalf("expected configuration_resource_id null")
	}
	if !state.InlineConfiguration.IsNull() {
		t.Fatalf("expected inline_configuration null")
	}
}
