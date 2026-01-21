package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestClient(rt http.RoundTripper) *APIClient {
	return &APIClient{
		httpClient: &http.Client{Transport: rt},
		cfg: providerConfig{
			domain:        "example.com",
			defaultRegion: "eu",
			apiKeySecret:  "secret",
		},
	}
}

func TestNewAPIClient_Defaults(t *testing.T) {
	c := NewAPIClient(providerConfig{domain: "example.com", defaultRegion: "eu", apiKeySecret: "secret"})
	if c.httpClient == nil {
		t.Fatalf("expected http client")
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Fatalf("unexpected timeout: %s", c.httpClient.Timeout)
	}
}

func httpResponse(status int, headers map[string]string, body string) *http.Response {
	h := make(http.Header)
	for k, v := range headers {
		h.Set(k, v)
	}
	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestAPIClient_doJSONInternal_SuccessJSONDecode(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("unexpected Authorization header: %q", got)
		}
		return httpResponse(200, nil, `{"ok":true}`), nil
	}))

	var out struct {
		Ok bool `json:"ok"`
	}
	found, status, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, &out, false)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !found || status != 200 || !out.Ok {
		t.Fatalf("unexpected result: found=%v status=%d out=%+v", found, status, out)
	}
}

func TestAPIClient_doJSONInternal_SuccessEmptyBodyOutNil(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(204, nil, ""), nil
	}))

	found, status, diags := c.doJSONInternal(context.Background(), "DELETE", "https://example.com/test", nil, nil, false)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !found || status != 204 {
		t.Fatalf("unexpected result: found=%v status=%d", found, status)
	}
}

func TestAPIClient_doJSONInternal_JSONEncodeError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("http should not be called on encode error")
		return nil, nil
	}))

	in := map[string]any{"bad": func() {}}
	_, _, diags := c.doJSONInternal(context.Background(), "POST", "https://example.com/test", in, nil, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestAPIClient_doJSONInternal_InvalidURL(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("http should not be called on invalid URL")
		return nil, nil
	}))

	_, _, diags := c.doJSONInternal(context.Background(), "GET", ":// bad", nil, nil, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestAPIClient_doJSONInternal_NetworkError(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return nil, errors.New("network down")
	}))

	// Avoid real sleeps during retry.
	oldSleep := sleepFn
	sleepFn = func(time.Duration) {}
	defer func() { sleepFn = oldSleep }()

	_, _, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, nil, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
	if calls != 5 {
		t.Fatalf("expected 5 attempts, got %d", calls)
	}
}

func TestAPIClient_doJSONInternal_Non2xxFriendlyError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(400, nil, `{"code":"BadThing","message":"nope"}`), nil
	}))

	_, _, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, nil, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
	if got := diags.Errors()[0].Detail(); !strings.Contains(got, "nope") || !strings.Contains(got, "BadThing") {
		t.Fatalf("unexpected error detail: %q", got)
	}
}

func TestAPIClient_doJSONInternal_Non2xxNonFriendlyBodySurfaced(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(500, nil, "oops"), nil
	}))

	_, _, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, nil, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
	if got := diags.Errors()[0].Detail(); !strings.Contains(got, "status=500") || !strings.Contains(got, "body=oops") {
		t.Fatalf("unexpected error detail: %q", got)
	}
}

func TestAPIClient_doJSONInternal_RetryBehaviorAndRetryAfterSeconds(t *testing.T) {
	var mu sync.Mutex
	var calls int

	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		if calls < 3 {
			return httpResponse(429, map[string]string{"Retry-After": "2"}, `{"code":"TooMany","message":"slow down"}`), nil
		}
		return httpResponse(200, nil, `{"ok":true}`), nil
	}))

	oldSleep := sleepFn
	var sleeps []time.Duration
	sleepFn = func(d time.Duration) { sleeps = append(sleeps, d) }
	defer func() { sleepFn = oldSleep }()

	var out map[string]any
	found, status, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, &out, false)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if !found || status != 200 {
		t.Fatalf("unexpected result: found=%v status=%d", found, status)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	if len(sleeps) != 2 {
		t.Fatalf("expected 2 sleeps, got %d", len(sleeps))
	}
	if sleeps[0] != 2*time.Second || sleeps[1] != 2*time.Second {
		t.Fatalf("expected Retry-After sleeps of 2s, got %#v", sleeps)
	}
}

func TestAPIClient_doJSONAllowNotFound_404Allowed(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	found, diags := c.doJSONAllowNotFound(context.Background(), "GET", "https://example.com/test", nil, nil)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if found {
		t.Fatalf("expected found=false")
	}
}

func TestAPIClient_doJSONInternal_JSONDecodeError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(200, nil, "not-json"), nil
	}))

	var out map[string]any
	_, _, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, &out, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestAPIClient_doJSONInternal_RequestBuildError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("http should not be called on request build error")
		return nil, nil
	}))

	_, _, diags := c.doJSONInternal(context.Background(), "B AD", "https://example.com/test", nil, nil, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestAPIClient_doJSONInternal_404NotAllowedIsError(t *testing.T) {
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return httpResponse(404, nil, `{"code":"NotFound","message":"missing"}`), nil
	}))

	_, _, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, nil, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
}

func TestAPIClient_doJSONInternal_RetryAfterDateUsesNowFn(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	retryAt := start.Add(2 * time.Second)

	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return httpResponse(429, map[string]string{"Retry-After": retryAt.UTC().Format(http.TimeFormat)}, `{"code":"TooMany","message":"slow down"}`), nil
		}
		return httpResponse(200, nil, `{}`), nil
	}))

	oldSleep := sleepFn
	oldNow := nowFn
	var slept time.Duration
	sleepFn = func(d time.Duration) { slept = d }
	nowFn = func() time.Time { return start }
	t.Cleanup(func() {
		sleepFn = oldSleep
		nowFn = oldNow
	})

	_, _, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, nil, false)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
	if slept != 2*time.Second {
		t.Fatalf("unexpected sleep: %s", slept)
	}
}

func TestAPIClient_doJSONInternal_RetryableStatusFallsBackToBackoffWhenRetryAfterBad(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 1 {
			return httpResponse(503, map[string]string{"Retry-After": "not-a-number"}, `{"code":"Down","message":"nope"}`), nil
		}
		return httpResponse(200, nil, `{}`), nil
	}))

	oldSleep := sleepFn
	var slept []time.Duration
	sleepFn = func(d time.Duration) { slept = append(slept, d) }
	t.Cleanup(func() { sleepFn = oldSleep })

	_, _, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, nil, false)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %#v", diags)
	}
	if calls != 2 {
		t.Fatalf("expected retry then success, got %d calls", calls)
	}
	if len(slept) != 1 {
		t.Fatalf("expected 1 sleep, got %d", len(slept))
	}
	if slept[0] < 50*time.Millisecond || slept[0] > 5*time.Second {
		t.Fatalf("unexpected sleep duration: %s", slept[0])
	}
}

func TestAPIClient_doJSONInternal_RetryableStatusNoRetryAtMaxAttempt(t *testing.T) {
	var calls int
	c := newTestClient(roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return httpResponse(503, nil, "oops"), nil
	}))

	oldSleep := sleepFn
	var slept int
	sleepFn = func(time.Duration) { slept++ }
	t.Cleanup(func() { sleepFn = oldSleep })

	_, _, diags := c.doJSONInternal(context.Background(), "GET", "https://example.com/test", nil, nil, false)
	if !diags.HasError() {
		t.Fatalf("expected error diagnostics")
	}
	if calls != 5 {
		t.Fatalf("expected 5 attempts, got %d", calls)
	}
	if slept != 4 {
		// Should sleep after attempts 1-4, not after final attempt.
		t.Fatalf("expected 4 sleeps, got %d", slept)
	}
}
