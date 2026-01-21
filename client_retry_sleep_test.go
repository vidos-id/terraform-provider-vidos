package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestRetrySleep_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, ok := retrySleep(ctx, 1, "", time.Now()); ok {
		t.Fatalf("expected ok=false")
	}
}

func TestRetrySleep_RetryAfterDateUsesProvidedNow(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	retryAt := now.Add(2 * time.Second)

	d, ok := retrySleep(context.Background(), 1, retryAt.UTC().Format(http.TimeFormat), now)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if d != 2*time.Second {
		t.Fatalf("unexpected duration: %s", d)
	}
}

func TestRetrySleep_RetryAfterSecondsBadOrNonPositiveFallsBackToBackoff(t *testing.T) {
	for _, tc := range []string{"0", "-1", "not-a-number"} {
		d, ok := retrySleep(context.Background(), 1, tc, time.Now())
		if !ok {
			t.Fatalf("expected ok=true for %q", tc)
		}
		if d < 50*time.Millisecond || d > 5*time.Second {
			t.Fatalf("unexpected duration for %q: %s", tc, d)
		}
	}
}

func TestRetrySleep_RetryAfterDateZeroOrPastFallsBackToBackoff(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, retryAt := range []time.Time{time.Time{}, now.Add(-1 * time.Second)} {
		d, ok := retrySleep(context.Background(), 1, retryAt.UTC().Format(http.TimeFormat), now)
		if !ok {
			t.Fatalf("expected ok=true")
		}
		if d < 50*time.Millisecond || d > 5*time.Second {
			t.Fatalf("unexpected duration: %s", d)
		}
	}
}

func TestRetrySleep_CapsAtMaxAndHasFloor(t *testing.T) {
	// Use a high attempt to ensure max cap.
	d, ok := retrySleep(context.Background(), 100, "", time.Now())
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if d > 5*time.Second {
		t.Fatalf("expected cap at 5s, got %s", d)
	}
	if d < 50*time.Millisecond {
		t.Fatalf("expected floor at 50ms, got %s", d)
	}
}
