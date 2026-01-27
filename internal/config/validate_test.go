package config

import "testing"

func TestNormalizeMode(t *testing.T) {
	got, err := NormalizeMode("Auto")
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if got != "auto" {
		t.Fatalf("expected auto, got %q", got)
	}

	if _, err := NormalizeMode("nope"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseHostURL(t *testing.T) {
	if _, err := ParseHostURL("http://127.0.0.1:11434"); err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if _, err := ParseHostURL("127.0.0.1:11434"); err == nil {
		t.Fatalf("expected error for missing scheme")
	}
	if _, err := ParseHostURL("https://example.com/path"); err == nil {
		t.Fatalf("expected error for path")
	}
}
