package main

import (
	"context"
	"net/http"
	"testing"
)

func TestAPIClient_doJSON_CallsInternal(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(200, nil, `{"ok":true}`), nil
	}))

	var out map[string]any
	diags := c.doJSON(context.Background(), "GET", "https://example.com/test", nil, &out)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
}
