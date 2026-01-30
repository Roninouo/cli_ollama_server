package ollamaapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestClientVersion(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/version" {
			t.Errorf("expected /api/version, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"version":"0.1.0"}`)
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	v, err := c.Version(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "0.1.0" {
		t.Errorf("expected 0.1.0, got %s", v)
	}
}

func TestClientVersionEmptyError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"version":""}`)
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	_, err := c.Version(context.Background())
	if err == nil {
		t.Fatal("expected error for empty version")
	}
}

func TestClientTags(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Errorf("expected /api/tags, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"models":[{"name":"llama3:8b","digest":"%s","size":4000000000,"modified_at":"%s"}]}`,
			strings.Repeat("a", 64),
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		)
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	models, err := c.Tags(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	if models[0].Name != "llama3:8b" {
		t.Errorf("expected llama3:8b, got %s", models[0].Name)
	}
}

func TestClientPS(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ps" {
			t.Errorf("expected /api/ps, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"models":[{"name":"llama3:8b","model":"llama3:8b","size":1234}]}`)
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	models, err := c.PS(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
}

func TestClientShow(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" {
			t.Errorf("expected /api/show, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var req ShowRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "llama3:8b" {
			t.Errorf("expected llama3:8b, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"llama3:8b","license":"MIT"}`)
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	resp, err := c.Show(context.Background(), "llama3:8b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["name"] != "llama3:8b" {
		t.Errorf("expected name=llama3:8b, got %v", resp["name"])
	}
}

func TestClientShowEmptyName(t *testing.T) {
	u, _ := url.Parse("http://localhost:11434")
	c := NewClient(u, false)

	_, err := c.Show(context.Background(), "   ")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestClientGenerate(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Errorf("expected /api/generate, got %s", r.URL.Path)
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Model != "llama3:8b" {
			t.Errorf("expected model llama3:8b, got %s", req.Model)
		}
		if req.Prompt != "hello" {
			t.Errorf("expected prompt hello, got %s", req.Prompt)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"response":"Hi","done":false}`+"\n")
		fmt.Fprint(w, `{"response":"!","done":true}`+"\n")
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	var out strings.Builder
	err := c.Generate(context.Background(), GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "hello",
		Stream: true,
	}, &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != "Hi!" {
		t.Errorf("expected 'Hi!', got %q", out.String())
	}
}

func TestClientGenerateEmptyModel(t *testing.T) {
	u, _ := url.Parse("http://localhost:11434")
	c := NewClient(u, false)

	var out strings.Builder
	err := c.Generate(context.Background(), GenerateRequest{
		Model:  "",
		Prompt: "hello",
	}, &out)
	if err == nil {
		t.Fatal("expected error for empty model")
	}
}

func TestClientGenerateEmptyPrompt(t *testing.T) {
	u, _ := url.Parse("http://localhost:11434")
	c := NewClient(u, false)

	var out strings.Builder
	err := c.Generate(context.Background(), GenerateRequest{
		Model:  "llama3:8b",
		Prompt: "   ",
	}, &out)
	if err == nil {
		t.Fatal("expected error for empty prompt")
	}
}

func TestClientGenerateWithAPIError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"response":"","done":false,"error":"model not found"}`+"\n")
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	var out strings.Builder
	err := c.Generate(context.Background(), GenerateRequest{
		Model:  "nonexistent",
		Prompt: "hello",
	}, &out)
	if err == nil {
		t.Fatal("expected error for API error")
	}
	if !strings.Contains(err.Error(), "model not found") {
		t.Errorf("expected 'model not found' in error, got: %v", err)
	}
}

func TestClientPull(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pull" {
			t.Errorf("expected /api/pull, got %s", r.URL.Path)
		}

		var req PullRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "llama3:8b" {
			t.Errorf("expected name llama3:8b, got %s", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"pulling manifest"}`+"\n")
		fmt.Fprint(w, `{"status":"downloading","digest":"sha256:abc123","total":1000,"completed":500}`+"\n")
		fmt.Fprint(w, `{"status":"done"}`+"\n")
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	var out strings.Builder
	err := c.Pull(context.Background(), "llama3:8b", &out)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "pulling manifest") {
		t.Errorf("expected 'pulling manifest' in output, got: %s", out.String())
	}
}

func TestClientPullEmptyName(t *testing.T) {
	u, _ := url.Parse("http://localhost:11434")
	c := NewClient(u, false)

	var out strings.Builder
	err := c.Pull(context.Background(), "  ", &out)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestClientPullWithAPIError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"","error":"unauthorized"}`+"\n")
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	var out strings.Builder
	err := c.Pull(context.Background(), "llama3:8b", &out)
	if err == nil {
		t.Fatal("expected error for API error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("expected 'unauthorized' in error, got: %v", err)
	}
}

func TestClientHTTPError(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error":"internal server error"}`)
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	_, err := c.Version(context.Background())
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestClientConnectionError(t *testing.T) {
	u, _ := url.Parse("http://127.0.0.1:59999") // non-existent port
	c := NewClient(u, false)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := c.Version(ctx)
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestClientContextCancellation(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Slow response
		fmt.Fprint(w, `{"version":"0.1.0"}`)
	}))
	defer s.Close()

	u, _ := url.Parse(s.URL)
	c := NewClient(u, false)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := c.Version(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
