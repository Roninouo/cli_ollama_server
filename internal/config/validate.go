package config

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeMode validates and normalizes the requested execution mode.
//
// Allowed values: auto, wrapper, native.
func NormalizeMode(in string) (string, error) {
	v := strings.TrimSpace(strings.ToLower(in))
	if v == "" {
		v = "auto"
	}
	switch v {
	case "auto", "wrapper", "native":
		return v, nil
	default:
		return "", fmt.Errorf("invalid mode: %s", in)
	}
}

// ParseHostURL validates host as an absolute http(s) URL suitable as an API base.
func ParseHostURL(host string) (*url.URL, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}
	u, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("invalid host: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("invalid host scheme: %s", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("invalid host: missing host")
	}
	if u.User != nil {
		return nil, fmt.Errorf("invalid host: userinfo not allowed")
	}
	if u.Fragment != "" {
		return nil, fmt.Errorf("invalid host: fragment not allowed")
	}
	if u.RawQuery != "" {
		return nil, fmt.Errorf("invalid host: query not allowed")
	}
	if u.Path != "" && u.Path != "/" {
		return nil, fmt.Errorf("invalid host: path not allowed")
	}
	u.Path = ""
	u.RawPath = ""
	return u, nil
}
