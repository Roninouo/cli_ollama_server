package ollamaapi

import (
	"strings"
	"testing"
	"time"
)

func TestFormatTags(t *testing.T) {
	models := []TagModel{
		{
			Name:       "llama3:8b",
			Digest:     "abc123def456abc123def456abc123def456abc123def456abc123def456abc123",
			Size:       4000000000,
			ModifiedAt: time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			Name:       "mistral:latest",
			Digest:     "xyz789",
			Size:       2500000000,
			ModifiedAt: time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC),
		},
	}

	out := FormatTags(models)

	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header NAME, got: %s", out)
	}
	if !strings.Contains(out, "llama3:8b") {
		t.Errorf("expected llama3:8b, got: %s", out)
	}
	if !strings.Contains(out, "mistral:latest") {
		t.Errorf("expected mistral:latest, got: %s", out)
	}
	// Check digest is truncated
	if !strings.Contains(out, "abc123def456") {
		t.Errorf("expected truncated digest, got: %s", out)
	}
	// Check full digest is NOT shown
	if strings.Contains(out, "abc123def456abc123def456abc123") {
		t.Errorf("digest should be truncated")
	}
}

func TestFormatTagsEmpty(t *testing.T) {
	out := FormatTags([]TagModel{})
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header even for empty list, got: %s", out)
	}
}

func TestFormatPS(t *testing.T) {
	models := []PSModel{
		{
			Name:      "llama3:8b",
			Digest:    "abc123def456",
			Size:      4000000000,
			ExpiresAt: time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC),
		},
	}

	out := FormatPS(models)

	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header NAME, got: %s", out)
	}
	if !strings.Contains(out, "UNTIL") {
		t.Errorf("expected header UNTIL, got: %s", out)
	}
	if !strings.Contains(out, "llama3:8b") {
		t.Errorf("expected llama3:8b, got: %s", out)
	}
}

func TestFormatPSEmpty(t *testing.T) {
	out := FormatPS([]PSModel{})
	if !strings.Contains(out, "NAME") {
		t.Errorf("expected header even for empty list, got: %s", out)
	}
}

func TestShortDigest(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc123", "abc123"},
		{"abc123def456", "abc123def456"},
		{"abc123def456xyz789", "abc123def456"},
		{"", ""},
		{"   ", ""},
		{"  abc123def456xyz789  ", "abc123def456"},
	}

	for _, tt := range tests {
		got := shortDigest(tt.input)
		if got != tt.expected {
			t.Errorf("shortDigest(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFmtBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{4000000000, "3.7 GB"},
		{1099511627776, "1.0 TB"},
		{-1, "?"},
	}

	for _, tt := range tests {
		got := fmtBytes(tt.input)
		if got != tt.expected {
			t.Errorf("fmtBytes(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFmtTime(t *testing.T) {
	tests := []struct {
		input    time.Time
		expected string
	}{
		{time.Time{}, "-"},
		{time.Date(2026, 1, 15, 10, 30, 45, 0, time.UTC), "2026-01-15 10:30:45Z"},
		{time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC), "2026-12-31 23:59:59Z"},
	}

	for _, tt := range tests {
		got := fmtTime(tt.input)
		if got != tt.expected {
			t.Errorf("fmtTime(%v) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatTableAlignment(t *testing.T) {
	// Test that columns are properly aligned
	models := []TagModel{
		{Name: "a", Digest: "d1", Size: 100, ModifiedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "long-model-name", Digest: "d2", Size: 200, ModifiedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
	}

	out := FormatTags(models)
	lines := strings.Split(strings.TrimSpace(out), "\n")

	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (header + 2 rows), got %d", len(lines))
	}

	// All lines should have consistent formatting
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			t.Errorf("line %d is empty", i)
		}
	}
}
