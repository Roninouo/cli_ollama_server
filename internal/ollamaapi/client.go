package ollamaapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Connection and timeout constants for the Ollama API client.
const (
	// DefaultDialTimeout is the timeout for establishing a connection.
	DefaultDialTimeout = 10 * time.Second
	// DefaultKeepAlive is the keep-alive duration for connections.
	DefaultKeepAlive = 30 * time.Second
	// DefaultTLSHandshakeTimeout is the timeout for TLS handshake.
	DefaultTLSHandshakeTimeout = 10 * time.Second
	// DefaultIdleConnTimeout is the timeout for idle connections.
	DefaultIdleConnTimeout = 90 * time.Second
	// DefaultExpectContinueTimeout is the timeout for expect-continue.
	DefaultExpectContinueTimeout = 1 * time.Second
	// DefaultMaxIdleConns is the maximum number of idle connections.
	DefaultMaxIdleConns = 10
	// DefaultMaxIdleConnsPerHost is the maximum idle connections per host.
	DefaultMaxIdleConnsPerHost = 5
)

// APIError represents an error returned by the Ollama API.
type APIError struct {
	StatusCode int
	Status     string
	Message    string
	Endpoint   string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("ollama api %s: %s (status %d)", e.Endpoint, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("ollama api %s: %s", e.Endpoint, e.Status)
}

// IsAPIError returns true if err is an APIError.
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// GetAPIError extracts the APIError from err if present.
func GetAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}

type Client struct {
	base  *url.URL
	http  *http.Client
	retry RetryConfig
}

// ClientOption allows customizing the client configuration.
type ClientOption func(*clientConfig)

type clientConfig struct {
	dialTimeout           time.Duration
	keepalive             time.Duration
	tlsHandshakeTimeout   time.Duration
	idleConnTimeout       time.Duration
	expectContinueTimeout time.Duration
	maxIdleConns          int
	maxIdleConnsPerHost   int
	retry                 RetryConfig
}

// WithDialTimeout sets a custom dial timeout.
func WithDialTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) { c.dialTimeout = d }
}

// WithIdleConnTimeout sets a custom idle connection timeout.
func WithIdleConnTimeout(d time.Duration) ClientOption {
	return func(c *clientConfig) { c.idleConnTimeout = d }
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) ClientOption {
	return func(c *clientConfig) { c.maxIdleConns = n }
}

func NewClient(base *url.URL, noProxyAuto bool, opts ...ClientOption) *Client {
	cfg := &clientConfig{
		dialTimeout:           DefaultDialTimeout,
		keepalive:             DefaultKeepAlive,
		tlsHandshakeTimeout:   DefaultTLSHandshakeTimeout,
		idleConnTimeout:       DefaultIdleConnTimeout,
		expectContinueTimeout: DefaultExpectContinueTimeout,
		maxIdleConns:          DefaultMaxIdleConns,
		maxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
		retry:                 NoRetry, // Disabled by default for backward compatibility
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// Avoid http.Client.Timeout for streaming endpoints.
	t := &http.Transport{
		Proxy: proxyFunc(noProxyAuto, base),
		DialContext: (&net.Dialer{
			Timeout:   cfg.dialTimeout,
			KeepAlive: cfg.keepalive,
		}).DialContext,
		TLSHandshakeTimeout:   cfg.tlsHandshakeTimeout,
		IdleConnTimeout:       cfg.idleConnTimeout,
		ExpectContinueTimeout: cfg.expectContinueTimeout,
		MaxIdleConns:          cfg.maxIdleConns,
		MaxIdleConnsPerHost:   cfg.maxIdleConnsPerHost,
	}
	return &Client{base: base, http: &http.Client{Transport: t}, retry: cfg.retry}
}

func (c *Client) Version(ctx context.Context) (string, error) {
	u := c.endpoint("/api/version")
	var resp VersionResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &resp); err != nil {
		return "", fmt.Errorf("get version: %w", err)
	}
	if strings.TrimSpace(resp.Version) == "" {
		return "", errors.New("ollama api: empty version response")
	}
	return resp.Version, nil
}

func (c *Client) Tags(ctx context.Context) ([]TagModel, error) {
	u := c.endpoint("/api/tags")
	var resp TagsResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &resp); err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	return resp.Models, nil
}

func (c *Client) PS(ctx context.Context) ([]PSModel, error) {
	u := c.endpoint("/api/ps")
	var resp PSResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &resp); err != nil {
		return nil, fmt.Errorf("list running models: %w", err)
	}
	return resp.Models, nil
}

func (c *Client) Show(ctx context.Context, name string) (map[string]any, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("show: empty model name")
	}
	u := c.endpoint("/api/show")
	var resp map[string]any
	if err := c.doJSON(ctx, http.MethodPost, u, ShowRequest{Name: name}, &resp); err != nil {
		return nil, fmt.Errorf("show model %q: %w", name, err)
	}
	return resp, nil
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest, w io.Writer) error {
	req.Model = strings.TrimSpace(req.Model)
	if req.Model == "" {
		return errors.New("generate: empty model")
	}
	// Keep prompt as-is (data), but disallow nil/empty to avoid silent interactive behavior.
	if strings.TrimSpace(req.Prompt) == "" {
		return errors.New("generate: empty prompt")
	}
	u := c.endpoint("/api/generate")

	h, err := c.doStream(ctx, u, req)
	if err != nil {
		return fmt.Errorf("generate with model %q: %w", req.Model, err)
	}
	defer h.Body.Close()

	dec := json.NewDecoder(h.Body)
	for {
		var chunk GenerateChunk
		if err := dec.Decode(&chunk); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("generate stream decode: %w", err)
		}
		if chunk.Error != "" {
			return &APIError{StatusCode: 0, Message: chunk.Error, Endpoint: "/api/generate"}
		}
		if chunk.Response != "" {
			if _, err := io.WriteString(w, chunk.Response); err != nil {
				return fmt.Errorf("write response: %w", err)
			}
		}
		if chunk.Done {
			break
		}
	}
	return nil
}

func (c *Client) Pull(ctx context.Context, name string, w io.Writer) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("pull: empty model name")
	}
	u := c.endpoint("/api/pull")

	h, err := c.doStream(ctx, u, PullRequest{Name: name, Stream: true})
	if err != nil {
		return fmt.Errorf("pull model %q: %w", name, err)
	}
	defer h.Body.Close()

	dec := json.NewDecoder(h.Body)
	for {
		var chunk PullChunk
		if err := dec.Decode(&chunk); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("pull stream decode: %w", err)
		}
		if chunk.Error != "" {
			return &APIError{StatusCode: 0, Message: chunk.Error, Endpoint: "/api/pull"}
		}
		// Best-effort human output without trying to match full ollama progress UI.
		if chunk.Digest != "" && chunk.Total > 0 {
			fmt.Fprintf(w, "%s %s %d/%d\n", strings.TrimSpace(chunk.Status), chunk.Digest, chunk.Completed, chunk.Total)
		} else if strings.TrimSpace(chunk.Status) != "" {
			fmt.Fprintln(w, strings.TrimSpace(chunk.Status))
		}
	}
	return nil
}

// Delete removes a model from the Ollama server.
func (c *Client) Delete(ctx context.Context, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("delete: empty model name")
	}
	u := c.endpoint("/api/delete")
	if err := c.doJSON(ctx, http.MethodDelete, u, DeleteRequest{Name: name}, nil); err != nil {
		return fmt.Errorf("delete model %q: %w", name, err)
	}
	return nil
}

// Copy duplicates a model with a new name on the Ollama server.
func (c *Client) Copy(ctx context.Context, source, destination string) error {
	source = strings.TrimSpace(source)
	destination = strings.TrimSpace(destination)
	if source == "" {
		return errors.New("copy: empty source model name")
	}
	if destination == "" {
		return errors.New("copy: empty destination model name")
	}
	u := c.endpoint("/api/copy")
	if err := c.doJSON(ctx, http.MethodPost, u, CopyRequest{Source: source, Destination: destination}, nil); err != nil {
		return fmt.Errorf("copy model %q to %q: %w", source, destination, err)
	}
	return nil
}

func (c *Client) endpoint(path string) string {
	u := *c.base
	u.Path = path
	return u.String()
}

func (c *Client) doJSON(ctx context.Context, method, url string, req any, out any) error {
	var lastErr error

	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry with exponential backoff.
			backoff := calculateBackoff(attempt-1, c.retry)
			if err := sleepWithContext(ctx, backoff); err != nil {
				return lastErr // Return the original error, not the sleep cancellation
			}
		}

		err := c.doJSONOnce(ctx, method, url, req, out)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if the error isn't transient.
		if !IsRetryableError(err) {
			return err
		}
	}

	return lastErr
}

func (c *Client) doJSONOnce(ctx context.Context, method, url string, req any, out any) error {
	var body io.Reader
	if req != nil {
		b, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(b)
	}
	hreq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if req != nil {
		hreq.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(hreq)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeAPIError(resp, extractEndpoint(url))
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *Client) doStream(ctx context.Context, url string, req any) (*http.Response, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	hreq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(hreq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, decodeAPIError(resp, extractEndpoint(url))
	}
	return resp, nil
}

// extractEndpoint extracts the path from a URL string for error messages.
func extractEndpoint(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	return u.Path
}

func decodeAPIError(resp *http.Response, endpoint string) *APIError {
	var e struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&e)
	return &APIError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Message:    strings.TrimSpace(e.Error),
		Endpoint:   endpoint,
	}
}
