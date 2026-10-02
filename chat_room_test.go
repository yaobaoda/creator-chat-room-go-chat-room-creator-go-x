package main

import (
	"strings"
	"testing"
)

func TestDeliveryMessage(t *testing.T) {
	cases := []struct{ name, input, want string }{
		{"trim whitespace", "  new asset  ", "new asset"},
		{"keep internal spacing", "release 42", "release 42"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeMessage(tc.input)
			if got != tc.want {
				t.Fatalf("normalizeMessage(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func normalizeMessage(s string) string { return strings.TrimSpace(s) }
