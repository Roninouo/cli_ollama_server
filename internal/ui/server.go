package ui

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"cli_ollama_server/internal/config"
	"cli_ollama_server/internal/execollama"
	"cli_ollama_server/internal/i18n"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	Listener   net.Listener
	Translator *i18n.Bundle
	Effective  config.Effective
	ConfigPath string
	BaseEnv    []string

	srv   *http.Server
	token string
	once  sync.Once
	wait  chan error
}

func (s *Server) Start() (string, error) {
	if s.Listener == nil {
		return "", fmt.Errorf("missing listener")
	}
	if s.Translator == nil {
		s.Translator = i18n.New("en")
	}
	s.wait = make(chan error, 1)

	tok, err := randomToken(16)
	if err != nil {
		return "", err
	}
	s.token = tok

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.HandleFunc("/api/list", s.auth(s.handleList))
	mux.HandleFunc("/api/pull", s.auth(s.handlePull))
	mux.HandleFunc("/api/run", s.auth(s.handleRun))
	mux.HandleFunc("/api/config", s.auth(s.handleConfig))
	mux.HandleFunc("/api/config/set", s.auth(s.handleConfigSet))

	s.srv = &http.Server{Handler: mux}

	addr := s.Listener.Addr().String()
	url := "http://" + addr + "/?t=" + s.token

	go func() {
		s.wait <- s.srv.Serve(s.Listener)
	}()

	_ = openBrowser(url)
	return url, nil
}

func (s *Server) Wait() int {
	err := <-s.wait
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return 0
	}
	return 1
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := r.Header.Get("X-Token")
		if t == "" {
			t = r.URL.Query().Get("t")
		}
		if t == "" || t != s.token {
			respondErr(w, http.StatusUnauthorized, s.Translator.Sprintf("ui.error.unauthorized"))
			return
		}
		next(w, r)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	raw, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "missing index", http.StatusInternalServerError)
		return
	}
	html := string(raw)
	html = strings.ReplaceAll(html, "{{TOKEN}}", s.token)
	html = strings.ReplaceAll(html, "{{LANG}}", s.Translator.Lang())
	html = strings.ReplaceAll(html, "{{TITLE}}", "ollama-remote")

	html = strings.ReplaceAll(html, "{{SUBTITLE}}", s.Translator.Sprintf("ui.subtitle"))
	html = strings.ReplaceAll(html, "{{BTN_LIST}}", s.Translator.Sprintf("ui.btn.list"))

	html = strings.ReplaceAll(html, "{{SECTION_CONFIG}}", s.Translator.Sprintf("ui.section.config"))
	html = strings.ReplaceAll(html, "{{LABEL_HOST}}", s.Translator.Sprintf("ui.label.host"))
	html = strings.ReplaceAll(html, "{{LABEL_LANG}}", s.Translator.Sprintf("ui.label.language"))
	html = strings.ReplaceAll(html, "{{BTN_SAVE}}", s.Translator.Sprintf("ui.btn.save"))
	html = strings.ReplaceAll(html, "{{PLACEHOLDER_HOST}}", s.Translator.Sprintf("ui.placeholder.host"))

	html = strings.ReplaceAll(html, "{{SECTION_PULL}}", s.Translator.Sprintf("ui.section.pull"))
	html = strings.ReplaceAll(html, "{{BTN_PULL}}", s.Translator.Sprintf("ui.btn.pull"))
	html = strings.ReplaceAll(html, "{{PLACEHOLDER_MODEL}}", s.Translator.Sprintf("ui.placeholder.model"))

	html = strings.ReplaceAll(html, "{{SECTION_RUN}}", s.Translator.Sprintf("ui.section.run"))
	html = strings.ReplaceAll(html, "{{LABEL_MODEL}}", s.Translator.Sprintf("ui.label.model"))
	html = strings.ReplaceAll(html, "{{LABEL_PROMPT}}", s.Translator.Sprintf("ui.label.prompt"))
	html = strings.ReplaceAll(html, "{{BTN_RUN}}", s.Translator.Sprintf("ui.btn.run"))
	html = strings.ReplaceAll(html, "{{PLACEHOLDER_PROMPT}}", s.Translator.Sprintf("ui.placeholder.prompt"))

	html = strings.ReplaceAll(html, "{{SECTION_OUTPUT}}", s.Translator.Sprintf("ui.section.output"))
	html = strings.ReplaceAll(html, "{{MSG_SAVED_RESTART}}", s.Translator.Sprintf("ui.message.saved_restart"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, html)
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	out, code, err := s.runOllama([]string{"list"})
	respondExec(w, out, code, err)
}

func (s *Server) handlePull(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, s.Translator.Sprintf("ui.error.bad_request"))
		return
	}
	req.Model = strings.TrimSpace(req.Model)
	if req.Model == "" {
		respondErr(w, http.StatusBadRequest, s.Translator.Sprintf("ui.error.model_required"))
		return
	}
	out, code, err := s.runOllama([]string{"pull", req.Model})
	respondExec(w, out, code, err)
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, s.Translator.Sprintf("ui.error.bad_request"))
		return
	}
	req.Model = strings.TrimSpace(req.Model)
	if req.Model == "" {
		respondErr(w, http.StatusBadRequest, s.Translator.Sprintf("ui.error.model_required"))
		return
	}
	out, code, err := s.runOllama([]string{"run", req.Model, req.Prompt})
	respondExec(w, out, code, err)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"configPath": s.ConfigPath,
		"host":       s.Effective.Host,
		"lang":       s.Translator.Lang(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleConfigSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Host string `json:"host"`
		Lang string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, s.Translator.Sprintf("ui.error.bad_request"))
		return
	}
	if strings.TrimSpace(req.Host) != "" {
		if err := config.SetUserConfig(s.ConfigPath, "host", strings.TrimSpace(req.Host)); err != nil {
			respondErr(w, http.StatusBadRequest, err.Error())
			return
		}
		s.Effective.Host = strings.TrimSpace(req.Host)
	}
	if strings.TrimSpace(req.Lang) != "" {
		if err := config.SetUserConfig(s.ConfigPath, "lang", strings.TrimSpace(req.Lang)); err != nil {
			respondErr(w, http.StatusBadRequest, err.Error())
			return
		}
		s.Translator = i18n.New(strings.TrimSpace(req.Lang))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) runOllama(args []string) (string, int, error) {
	env, _, _ := config.BuildChildEnv(config.ChildEnvOptions{Existing: s.BaseEnv, Effective: s.Effective})
	var b strings.Builder
	code, err := execollama.Run(execollama.RunOptions{
		Args:      args,
		Env:       env,
		OllamaExe: s.Effective.OllamaExe,
		Stdout:    &b,
		Stderr:    &b,
	})
	if err != nil && errors.Is(err, execollama.ErrNotFound) {
		return b.String(), code, execollama.ErrNotFound
	}
	return b.String(), code, err
}

func respondExec(w http.ResponseWriter, out string, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{"output": out, "exitCode": code}
	if err != nil {
		resp["error"] = err.Error()
	}
	json.NewEncoder(w).Encode(resp)
}

func respondErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"error": msg, "exitCode": status})
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
