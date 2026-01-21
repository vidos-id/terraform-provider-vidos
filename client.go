package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type APIClient struct {
	httpClient *http.Client
	cfg        providerConfig
}

// sleepFn exists to make retry behavior unit-testable without real delays.
// Production code uses time.Sleep.
var sleepFn = time.Sleep

// nowFn exists to make time-dependent behavior unit-testable.
// Production code uses time.Now.
var nowFn = time.Now

func NewAPIClient(cfg providerConfig) *APIClient {
	return &APIClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		cfg:        cfg,
	}
}

type friendlyError struct {
	Code    string `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
	Action  string `json:"action"`
}

type apiError struct {
	StatusCode int
	Method     string
	URL        string
	Body       string
	Friendly   *friendlyError
}

func (e apiError) Error() string {
	if e.Friendly != nil {
		code := strings.TrimSpace(e.Friendly.Code)
		if code == "" {
			code = strings.TrimSpace(e.Friendly.Type)
		}
		if code != "" {
			return fmt.Sprintf("%s %s failed: %s (%s)", e.Method, e.URL, e.Friendly.Message, code)
		}
		return fmt.Sprintf("%s %s failed: %s", e.Method, e.URL, e.Friendly.Message)
	}
	if strings.TrimSpace(e.Body) != "" {
		return fmt.Sprintf("%s %s failed: status=%d body=%s", e.Method, e.URL, e.StatusCode, truncateForError(e.Body, 1024))
	}
	return fmt.Sprintf("%s %s failed: status=%d", e.Method, e.URL, e.StatusCode)
}

func (c *APIClient) iamBaseURL() string {
	return buildManagementBaseURL("iam", "global", c.cfg.domain)
}

func (c *APIClient) resolverBaseURL() string {
	return buildManagementBaseURL("resolver", c.cfg.defaultRegion, c.cfg.domain)
}

func (c *APIClient) verifierBaseURL() string {
	return buildManagementBaseURL("verifier", c.cfg.defaultRegion, c.cfg.domain)
}

func (c *APIClient) validatorBaseURL() string {
	return buildManagementBaseURL("validator", c.cfg.defaultRegion, c.cfg.domain)
}

func (c *APIClient) authorizerBaseURL() string {
	return buildManagementBaseURL("authorizer", c.cfg.defaultRegion, c.cfg.domain)
}

func (c *APIClient) gatewayBaseURL() string {
	return buildManagementBaseURL("gateway", c.cfg.defaultRegion, c.cfg.domain)
}

func (c *APIClient) doJSON(ctx context.Context, method, rawURL string, in any, out any) diag.Diagnostics {
	_, _, diags := c.doJSONInternal(ctx, method, rawURL, in, out, false)
	return diags
}

func (c *APIClient) doJSONAllowNotFound(ctx context.Context, method, rawURL string, in any, out any) (bool, diag.Diagnostics) {
	found, _, diags := c.doJSONInternal(ctx, method, rawURL, in, out, true)
	return found, diags
}

func (c *APIClient) doJSONInternal(ctx context.Context, method, rawURL string, in any, out any, allowNotFound bool) (bool, int, diag.Diagnostics) {
	var diags diag.Diagnostics

	u, err := url.Parse(rawURL)
	if err != nil {
		diags.AddError("Invalid URL", err.Error())
		return false, 0, diags
	}

	var bodyBytes []byte
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			diags.AddError("JSON encode error", err.Error())
			return false, 0, diags
		}
		bodyBytes = b
	}

	const maxAttempts = 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var body io.Reader
		if bodyBytes != nil {
			body = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
		if err != nil {
			diags.AddError("Request build error", err.Error())
			return false, 0, diags
		}
		req.Header.Set("Authorization", "Bearer "+c.cfg.apiKeySecret)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "terraform-provider-vidos")
		req.Header.Set("X-Vidos-Api-Version", "1.0")
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt < maxAttempts {
				if sleep, ok := retrySleep(ctx, attempt, "", time.Time{}); ok {
					sleepFn(sleep)
					continue
				}
			}
			diags.AddError("Request error", err.Error())
			return false, 0, diags
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			if allowNotFound && resp.StatusCode == 404 {
				return false, resp.StatusCode, diags
			}

			var fe friendlyError
			_ = json.Unmarshal(respBody, &fe)
			isFriendly := strings.TrimSpace(fe.Message) != "" && (strings.TrimSpace(fe.Code) != "" || strings.TrimSpace(fe.Type) != "")
			if !isFriendly {
				fe = friendlyError{}
			}

			retryable := resp.StatusCode == 429 || resp.StatusCode == 502 || resp.StatusCode == 503 || resp.StatusCode == 504
			if retryable && attempt < maxAttempts {
				retryAfter := resp.Header.Get("Retry-After")
				if sleep, ok := retrySleep(ctx, attempt, retryAfter, nowFn()); ok {
					tflog.Debug(ctx, "Retrying request", map[string]any{"attempt": attempt, "status": resp.StatusCode, "sleep": sleep.String(), "url": u.String()})
					sleepFn(sleep)
					continue
				}
			}

			if fe.Message != "" {
				diags.AddError("API error", apiError{StatusCode: resp.StatusCode, Method: method, URL: u.String(), Body: string(respBody), Friendly: &fe}.Error())
				return false, resp.StatusCode, diags
			}
			diags.AddError("API error", apiError{StatusCode: resp.StatusCode, Method: method, URL: u.String(), Body: string(respBody)}.Error())
			return false, resp.StatusCode, diags
		}

		if out == nil || len(respBody) == 0 {
			return true, resp.StatusCode, diags
		}

		if err := json.Unmarshal(respBody, out); err != nil {
			tflog.Debug(ctx, "Response body", map[string]any{"body": string(respBody)})
			diags.AddError("JSON decode error", err.Error())
			return false, resp.StatusCode, diags
		}

		return true, resp.StatusCode, diags
	}

	return false, 0, diags
}

func truncateForError(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func retrySleep(ctx context.Context, attempt int, retryAfter string, now time.Time) (time.Duration, bool) {
	if ctx.Err() != nil {
		return 0, false
	}
	if retryAfter != "" {
		if secs, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second, true
		}
		if t, err := http.ParseTime(retryAfter); err == nil && !t.IsZero() && !now.IsZero() {
			d := t.Sub(now)
			if d > 0 {
				return d, true
			}
		}
	}

	// Exponential backoff with jitter. Capped to keep Terraform responsive.
	base := 250 * time.Millisecond
	max := 5 * time.Second
	// attempt starts at 1
	sleep := base * time.Duration(1<<(attempt-1))
	if sleep > max {
		sleep = max
	}
	// jitter in [0.5, 1.5)
	jitter := 0.5 + rand.Float64()
	sleep = time.Duration(float64(sleep) * jitter)
	if sleep < 50*time.Millisecond {
		sleep = 50 * time.Millisecond
	}
	return sleep, true
}
