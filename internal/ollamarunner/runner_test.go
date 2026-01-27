package ollamarunner

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cli_ollama_server/internal/i18n"
)

func TestNativeListAndRun(t *testing.T) {
	s := newFakeOllamaServer(t)
	defer s.Close()

	tr := i18n.New("en")

	{
		var out strings.Builder
		code, err := Run(context.Background(), Options{
			Mode:       "native",
			Host:       s.URL,
			Args:       []string{"list"},
			Stdout:     &out,
			Stderr:     &out,
			Translator: tr,
		})
		if err != nil || code != 0 {
			t.Fatalf("list: code=%d err=%v out=%q", code, err, out.String())
		}
		if !strings.Contains(out.String(), "NAME") {
			t.Fatalf("expected table header, got %q", out.String())
		}
	}

	{
		var out strings.Builder
		code, err := Run(context.Background(), Options{
			Mode:       "native",
			Host:       s.URL,
			Args:       []string{"run", "llama3:8b", "--", "hello"},
			Stdout:     &out,
			Stderr:     &out,
			Stdin:      strings.NewReader(""),
			Translator: tr,
		})
		if err != nil || code != 0 {
			t.Fatalf("run: code=%d err=%v out=%q", code, err, out.String())
		}
		if out.String() != "hi!" {
			t.Fatalf("expected streamed output, got %q", out.String())
		}
	}
}

func TestNativePullRequiresUnsafe(t *testing.T) {
	s := newFakeOllamaServer(t)
	defer s.Close()

	tr := i18n.New("en")
	var out strings.Builder
	code, err := Run(context.Background(), Options{
		Mode:       "native",
		Host:       s.URL,
		Args:       []string{"pull", "llama3:8b"},
		Stdout:     &out,
		Stderr:     &out,
		Translator: tr,
	})
	if code != 2 {
		t.Fatalf("expected code=2, got %d", code)
	}
	if err == nil {
		t.Fatalf("expected error")
	}
}

func newFakeOllamaServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"version":"0.0.1"}`)
	})
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"models":[{"name":"llama3:8b","digest":"%s","size":1234,"modified_at":"%s"}]}`,
			strings.Repeat("a", 64),
			time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		)
	})
	mux.HandleFunc("/api/ps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"models":[]}`)
	})
	mux.HandleFunc("/api/show", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"llama3:8b"}`)
	})
	mux.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{\"response\":\"hi\",\"done\":false}\n")
		fmt.Fprint(w, "{\"response\":\"!\",\"done\":true}\n")
	})
	mux.HandleFunc("/api/pull", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{\"status\":\"pulling\",\"completed\":1,\"total\":2}\n")
		fmt.Fprint(w, "{\"status\":\"done\"}\n")
	})

	return httptest.NewServer(mux)
}
