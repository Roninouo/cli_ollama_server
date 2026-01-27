package ollamaapi

import (
	"net/http"
	"net/url"
	"strings"
)

func proxyFunc(noProxyAuto bool, base *url.URL) func(*http.Request) (*url.URL, error) {
	if !noProxyAuto || base == nil {
		return http.ProxyFromEnvironment
	}
	baseHost := strings.ToLower(strings.TrimSpace(base.Hostname()))
	if baseHost == "" {
		return http.ProxyFromEnvironment
	}
	return func(r *http.Request) (*url.URL, error) {
		h := strings.ToLower(strings.TrimSpace(r.URL.Hostname()))
		if h != "" && h == baseHost {
			// Bypass proxies for the configured host without mutating NO_PROXY.
			return nil, nil
		}
		return http.ProxyFromEnvironment(r)
	}
}
