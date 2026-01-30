package ollamaapi

import (
	"net/http"
	"net/url"
	"testing"
)

func TestProxyFunc(t *testing.T) {
	tests := []struct {
		name         string
		noProxyAuto  bool
		baseHost     string
		requestHost  string
		expectBypass bool
	}{
		{
			name:         "noProxyAuto=false returns ProxyFromEnvironment",
			noProxyAuto:  false,
			baseHost:     "http://localhost:11434",
			requestHost:  "http://localhost:11434/api/version",
			expectBypass: false, // Uses env proxy
		},
		{
			name:         "noProxyAuto=true bypasses proxy for same host",
			noProxyAuto:  true,
			baseHost:     "http://localhost:11434",
			requestHost:  "http://localhost:11434/api/version",
			expectBypass: true,
		},
		{
			name:         "noProxyAuto=true does not bypass for different host",
			noProxyAuto:  true,
			baseHost:     "http://localhost:11434",
			requestHost:  "http://other.example.com/api/version",
			expectBypass: false,
		},
		{
			name:         "case insensitive host matching",
			noProxyAuto:  true,
			baseHost:     "http://LocalHost:11434",
			requestHost:  "http://LOCALHOST:11434/api/version",
			expectBypass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, _ := url.Parse(tt.baseHost)
			pf := proxyFunc(tt.noProxyAuto, baseURL)

			req, _ := http.NewRequest("GET", tt.requestHost, nil)
			proxyURL, err := pf(req)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectBypass {
				if proxyURL != nil {
					t.Errorf("expected nil proxy (bypass), got %v", proxyURL)
				}
			}
			// Note: when expectBypass is false, the result depends on environment
			// so we can't easily verify, but at least we check no error
		})
	}
}

func TestProxyFuncNilBase(t *testing.T) {
	// nil base should return ProxyFromEnvironment
	pf := proxyFunc(true, nil)
	if pf == nil {
		t.Fatal("expected non-nil proxy func")
	}
}

func TestProxyFuncEmptyHost(t *testing.T) {
	baseURL, _ := url.Parse("http:///path") // empty host
	pf := proxyFunc(true, baseURL)
	if pf == nil {
		t.Fatal("expected non-nil proxy func")
	}
}
