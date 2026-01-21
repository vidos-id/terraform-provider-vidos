package main

import "testing"

func TestAPIError_Error_Friendly_UsesTypeWhenCodeEmpty(t *testing.T) {
	err := apiError{
		Method: "GET",
		URL:    "https://example.invalid/test",
		Friendly: &friendlyError{
			Code:    "",
			Type:    "NotFound",
			Message: "missing",
		},
	}

	if got := err.Error(); got != "GET https://example.invalid/test failed: missing (NotFound)" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestAPIError_Error_Friendly_OmitsParenWhenNoCodeOrType(t *testing.T) {
	err := apiError{
		Method: "POST",
		URL:    "https://example.invalid/test",
		Friendly: &friendlyError{
			Code:    "",
			Type:    "",
			Message: "nope",
		},
	}

	if got := err.Error(); got != "POST https://example.invalid/test failed: nope" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestAPIError_Error_NonFriendly_IncludesBodyWhenPresent(t *testing.T) {
	err := apiError{
		StatusCode: 503,
		Method:     "GET",
		URL:        "https://example.invalid/test",
		Body:       "service unavailable",
	}

	if got := err.Error(); got != "GET https://example.invalid/test failed: status=503 body=service unavailable" {
		t.Fatalf("unexpected error: %q", got)
	}
}

func TestAPIError_Error_NonFriendly_OmitsBodyWhenEmpty(t *testing.T) {
	err := apiError{
		StatusCode: 500,
		Method:     "GET",
		URL:        "https://example.invalid/test",
		Body:       "   ",
	}

	if got := err.Error(); got != "GET https://example.invalid/test failed: status=500" {
		t.Fatalf("unexpected error: %q", got)
	}
}
