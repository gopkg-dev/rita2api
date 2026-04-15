package upstream

import (
	"strings"
	"testing"
)

func TestBuildVisitorIDHeaderReturnsSignedValue(t *testing.T) {
	t.Parallel()

	header, err := BuildVisitorIDHeader("visitor-123", "secret-key")
	if err != nil {
		t.Fatalf("BuildVisitorIDHeader() error = %v", err)
	}

	parts := strings.Split(header, ":")
	if len(parts) != 2 {
		t.Fatalf("header format = %q, want visitor:signature", header)
	}

	if parts[0] != "visitor-123" {
		t.Fatalf("visitor part = %q, want %q", parts[0], "visitor-123")
	}

	if len(parts[1]) != 64 {
		t.Fatalf("signature length = %d, want 64", len(parts[1]))
	}
}

func TestBuildVisitorIDHeaderRejectsBlankInput(t *testing.T) {
	t.Parallel()

	if _, err := BuildVisitorIDHeader("", "secret-key"); err == nil {
		t.Fatal("BuildVisitorIDHeader() error = nil, want error for blank visitor id")
	}

	if _, err := BuildVisitorIDHeader("visitor-123", ""); err == nil {
		t.Fatal("BuildVisitorIDHeader() error = nil, want error for blank secret")
	}
}
