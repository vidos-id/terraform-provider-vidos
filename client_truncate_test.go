package main

import "testing"

func TestTruncateForError(t *testing.T) {
	{
		got := truncateForError("  hi  ", 10)
		if got != "hi" {
			t.Fatalf("unexpected: %q", got)
		}
	}
	{
		got := truncateForError("abcdef", 3)
		if got != "abc…" {
			t.Fatalf("unexpected: %q", got)
		}
	}
}
