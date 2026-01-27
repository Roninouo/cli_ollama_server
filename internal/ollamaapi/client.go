package ollamaapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	base *url.URL
	http *http.Client
}

func NewClient(base *url.URL, noProxyAuto bool) *Client {
	// Avoid http.Client.Timeout for streaming endpoints.
	t := &http.Transport{
		Proxy: proxyFunc(noProxyAuto, base),
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &Client{base: base, http: &http.Client{Transport: t}}
}

func (c *Client) Version(ctx context.Context) (string, error) {
	u := c.endpoint("/api/version")
	var resp VersionResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &resp); err != nil {
		return "", err
	}
	if strings.TrimSpace(resp.Version) == "" {
		return "", fmt.Errorf("empty version")
	}
	return resp.Version, nil
}

func (c *Client) Tags(ctx context.Context) ([]TagModel, error) {
	u := c.endpoint("/api/tags")
	var resp TagsResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
}

func (c *Client) PS(ctx context.Context) ([]PSModel, error) {
	u := c.endpoint("/api/ps")
	var resp PSResponse
	if err := c.doJSON(ctx, http.MethodGet, u, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Models, nil
}

func (c *Client) Show(ctx context.Context, name string) (map[string]any, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("empty model name")
	}
	u := c.endpoint("/api/show")
	var resp map[string]any
	if err := c.doJSON(ctx, http.MethodPost, u, ShowRequest{Name: name}, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest, w io.Writer) error {
	req.Model = strings.TrimSpace(req.Model)
	if req.Model == "" {
		return fmt.Errorf("empty model")
	}
	// Keep prompt as-is (data), but disallow nil/empty to avoid silent interactive behavior.
	if strings.TrimSpace(req.Prompt) == "" {
		return fmt.Errorf("empty prompt")
	}
	u := c.endpoint("/api/generate")

	h, err := c.doStream(ctx, u, req)
	if err != nil {
		return err
	}
	defer h.Body.Close()

	dec := json.NewDecoder(h.Body)
	for {
		var chunk GenerateChunk
		if err := dec.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if chunk.Error != "" {
			return fmt.Errorf("%s", chunk.Error)
		}
		if chunk.Response != "" {
			if _, err := io.WriteString(w, chunk.Response); err != nil {
				return err
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
		return fmt.Errorf("empty model name")
	}
	u := c.endpoint("/api/pull")

	h, err := c.doStream(ctx, u, PullRequest{Name: name, Stream: true})
	if err != nil {
		return err
	}
	defer h.Body.Close()

	dec := json.NewDecoder(h.Body)
	for {
		var chunk PullChunk
		if err := dec.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if chunk.Error != "" {
			return fmt.Errorf("%s", chunk.Error)
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

func (c *Client) endpoint(path string) string {
	u := *c.base
	u.Path = path
	return u.String()
}

func (c *Client) doJSON(ctx context.Context, method, url string, req any, out any) error {
	var body io.Reader
	if req != nil {
		b, err := json.Marshal(req)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	hreq, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return err
	}
	if req != nil {
		hreq.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(hreq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return decodeAPIError(resp)
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) doStream(ctx context.Context, url string, req any) (*http.Response, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	hreq.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(hreq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, decodeAPIError(resp)
	}
	return resp, nil
}

func decodeAPIError(resp *http.Response) error {
	var e struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&e)
	msg := strings.TrimSpace(e.Error)
	if msg == "" {
		msg = resp.Status
	}
	return fmt.Errorf("%s", msg)
}
