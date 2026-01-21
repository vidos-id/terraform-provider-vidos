package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestIamPolicyResource_ReadIntoState_UnsupportedPolicyType(t *testing.T) {
	r := &IamPolicyResource{}
	r.client = newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return httpResponse(200, nil, `{"policy":{"resourceId":"rid","name":"n","document":{},"policyType":"managed"}}`), nil
	}))

	var state iamPolicyModel
	found, diags := r.readIntoState(context.Background(), "rid", &state)
	if !found {
		t.Fatalf("expected found")
	}
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestIamPolicyResource_ReadIntoState_DocumentMarshalError(t *testing.T) {
	oldMarshal := jsonMarshal
	jsonMarshal = func(any) ([]byte, error) { return nil, errors.New("nope") }
	t.Cleanup(func() { jsonMarshal = oldMarshal })

	r := &IamPolicyResource{}
	r.client = newTestClient(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return httpResponse(200, nil, `{"policy":{"resourceId":"rid","name":"n","document":{"a":1},"policyType":"account"}}`), nil
	}))

	var state iamPolicyModel
	found, diags := r.readIntoState(context.Background(), "rid", &state)
	if !found {
		t.Fatalf("expected found")
	}
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}
