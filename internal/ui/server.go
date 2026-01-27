package ui

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"cli_ollama_server/internal/config"
	"cli_ollama_server/internal/execollama"
	"cli_ollama_server/internal/i18n"
	"cli_ollama_server/internal/ollamarunner"
)

//go:generate npx --yes esbuild ./frontend/app.ts --bundle --platform=browser --format=esm --target=es2020 --outfile=./static/app.js

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
	assets, err := fs.Sub(staticFS, "static")
	if err != nil {
		assets = staticFS
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(assets))))
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
	tr := s.Translator

	repl := map[string]string{
		"{{TOKEN}}": s.token,
		"{{LANG}}":  tr.Lang(),
		"{{TITLE}}": "ollama-remote",

		"{{SUBTITLE}}": tr.Sprintf("ui.subtitle"),

		"{{BTN_LIST}}": tr.Sprintf("ui.btn.list"),
		"{{BTN_SAVE}}": tr.Sprintf("ui.btn.save"),
		"{{BTN_PULL}}": tr.Sprintf("ui.btn.pull"),
		"{{BTN_RUN}}":  tr.Sprintf("ui.btn.run"),

		"{{SECTION_CONFIG}}":    tr.Sprintf("ui.section.config"),
		"{{SECTION_PULL}}":      tr.Sprintf("ui.section.pull"),
		"{{SECTION_RUN}}":       tr.Sprintf("ui.section.run"),
		"{{SECTION_OUTPUT}}":    tr.Sprintf("ui.section.output"),
		"{{UI_SECTION_MODELS}}": tr.Sprintf("ui.section.models"),

		"{{LABEL_HOST}}":   tr.Sprintf("ui.label.host"),
		"{{LABEL_LANG}}":   tr.Sprintf("ui.label.language"),
		"{{LABEL_MODEL}}":  tr.Sprintf("ui.label.model"),
		"{{LABEL_PROMPT}}": tr.Sprintf("ui.label.prompt"),

		"{{PLACEHOLDER_HOST}}":   tr.Sprintf("ui.placeholder.host"),
		"{{PLACEHOLDER_MODEL}}":  tr.Sprintf("ui.placeholder.model"),
		"{{PLACEHOLDER_PROMPT}}": tr.Sprintf("ui.placeholder.prompt"),

		"{{MSG_SAVED_RESTART}}": tr.Sprintf("ui.message.saved_restart"),

		"{{UI_NAV_MODELS}}":   tr.Sprintf("ui.nav.models"),
		"{{UI_NAV_RUN}}":      tr.Sprintf("ui.nav.run"),
		"{{UI_NAV_PULL}}":     tr.Sprintf("ui.nav.pull"),
		"{{UI_NAV_SETTINGS}}": tr.Sprintf("ui.nav.settings"),

		"{{UI_LABEL_SEARCH}}":           tr.Sprintf("ui.label.search"),
		"{{UI_PLACEHOLDER_SEARCH}}":     tr.Sprintf("ui.placeholder.search"),
		"{{UI_TABLE_MODEL}}":            tr.Sprintf("ui.table.model"),
		"{{UI_TABLE_ID}}":               tr.Sprintf("ui.table.id"),
		"{{UI_TABLE_SIZE}}":             tr.Sprintf("ui.table.size"),
		"{{UI_TABLE_MODIFIED}}":         tr.Sprintf("ui.table.modified"),
		"{{UI_TABLE_ACTIONS}}":          tr.Sprintf("ui.table.actions"),
		"{{UI_MODELS_HINT}}":            tr.Sprintf("ui.models.hint"),
		"{{UI_HINT_RUN_SHORTCUT}}":      tr.Sprintf("ui.hint.run_shortcut"),
		"{{UI_HINT_PULL_UNSAFE}}":       tr.Sprintf("ui.hint.pull_unsafe"),
		"{{UI_LABEL_CONFIG_PATH}}":      tr.Sprintf("ui.label.config_path"),
		"{{UI_LABEL_MODE}}":             tr.Sprintf("ui.label.mode"),
		"{{UI_MODE_AUTO}}":              tr.Sprintf("ui.mode.auto"),
		"{{UI_MODE_WRAPPER}}":           tr.Sprintf("ui.mode.wrapper"),
		"{{UI_MODE_NATIVE}}":            tr.Sprintf("ui.mode.native"),
		"{{UI_HINT_MODE}}":              tr.Sprintf("ui.hint.mode"),
		"{{UI_LABEL_OLLAMA_EXE}}":       tr.Sprintf("ui.label.ollama_exe"),
		"{{UI_PLACEHOLDER_OLLAMA_EXE}}": tr.Sprintf("ui.placeholder.ollama_exe"),
		"{{UI_LABEL_UNSAFE}}":           tr.Sprintf("ui.label.unsafe"),
		"{{UI_LABEL_NO_PROXY_AUTO}}":    tr.Sprintf("ui.label.no_proxy_auto"),

		"{{UI_BTN_WRAP}}":     tr.Sprintf("ui.btn.wrap"),
		"{{UI_BTN_UNWRAP}}":   tr.Sprintf("ui.btn.unwrap"),
		"{{UI_BTN_COPY}}":     tr.Sprintf("ui.btn.copy"),
		"{{UI_BTN_CLEAR}}":    tr.Sprintf("ui.btn.clear"),
		"{{UI_OUTPUT_EMPTY}}": tr.Sprintf("ui.output.empty"),

		"{{UI_MSG_COPIED}}":  tr.Sprintf("ui.message.copied"),
		"{{UI_MSG_LOADING}}": tr.Sprintf("ui.message.loading"),
		"{{UI_MSG_WORKING}}": tr.Sprintf("ui.message.working"),

		"{{UI_BTN_THEME}}":         tr.Sprintf("ui.btn.theme"),
		"{{UI_A11Y_NAV}}":          tr.Sprintf("ui.a11y.nav"),
		"{{UI_A11Y_STATUS}}":       tr.Sprintf("ui.a11y.status"),
		"{{UI_A11Y_MODELS_TABLE}}": tr.Sprintf("ui.a11y.models_table"),
	}

	for k, v := range repl {
		html = strings.ReplaceAll(html, k, v)
	}
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
	// Use "--" to ensure prompts that start with '-' are not treated as flags by the wrapper CLI.
	out, code, err := s.runOllama([]string{"run", req.Model, "--", req.Prompt})
	respondExec(w, out, code, err)
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	selected := s.Effective.Mode
	if selected == "auto" {
		if _, err := execollama.ResolveExecutable(s.Effective.OllamaExe); err == nil {
			selected = "wrapper"
		} else {
			selected = "native"
		}
	}

	resp := map[string]any{
		"configPath":   s.ConfigPath,
		"host":         s.Effective.Host,
		"lang":         s.Translator.Lang(),
		"mode":         s.Effective.Mode,
		"unsafe":       s.Effective.Unsafe,
		"noProxyAuto":  s.Effective.NoProxyAuto,
		"ollamaExe":    s.Effective.OllamaExe,
		"selectedMode": selected,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleConfigSet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Host        *string `json:"host"`
		Lang        *string `json:"lang"`
		Mode        *string `json:"mode"`
		OllamaExe   *string `json:"ollamaExe"`
		Unsafe      *bool   `json:"unsafe"`
		NoProxyAuto *bool   `json:"noProxyAuto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, s.Translator.Sprintf("ui.error.bad_request"))
		return
	}
	if req.Host != nil {
		v := strings.TrimSpace(*req.Host)
		if v == "" {
			v = "http://127.0.0.1:11434"
		}
		if _, herr := config.ParseHostURL(v); herr != nil {
			respondErr(w, http.StatusBadRequest, s.Translator.Sprintf("error.invalid_host", "host", v, "error", herr.Error()))
			return
		}
		if err := config.SetUserConfig(s.ConfigPath, "host", v); err != nil {
			respondErr(w, http.StatusBadRequest, err.Error())
			return
		}
		s.Effective.Host = v
	}
	if req.Lang != nil {
		v := strings.TrimSpace(*req.Lang)
		if v != "" {
			if err := config.SetUserConfig(s.ConfigPath, "lang", v); err != nil {
				respondErr(w, http.StatusBadRequest, err.Error())
				return
			}
			s.Translator = i18n.New(v)
		}
	}
	if req.Mode != nil {
		v := strings.TrimSpace(*req.Mode)
		if v == "" {
			v = "auto"
		}
		if m, merr := config.NormalizeMode(v); merr != nil {
			respondErr(w, http.StatusBadRequest, s.Translator.Sprintf("error.invalid_mode", "mode", v))
			return
		} else {
			v = m
		}
		if err := config.SetUserConfig(s.ConfigPath, "mode", v); err != nil {
			respondErr(w, http.StatusBadRequest, err.Error())
			return
		}
		s.Effective.Mode = v
	}
	if req.OllamaExe != nil {
		v := strings.TrimSpace(*req.OllamaExe)
		if err := config.SetUserConfig(s.ConfigPath, "ollama_exe", v); err != nil {
			respondErr(w, http.StatusBadRequest, err.Error())
			return
		}
		s.Effective.OllamaExe = v
	}
	if req.Unsafe != nil {
		v := "false"
		if *req.Unsafe {
			v = "true"
		}
		if err := config.SetUserConfig(s.ConfigPath, "unsafe", v); err != nil {
			respondErr(w, http.StatusBadRequest, err.Error())
			return
		}
		s.Effective.Unsafe = *req.Unsafe
	}
	if req.NoProxyAuto != nil {
		v := "false"
		if *req.NoProxyAuto {
			v = "true"
		}
		if err := config.SetUserConfig(s.ConfigPath, "no_proxy_auto", v); err != nil {
			respondErr(w, http.StatusBadRequest, err.Error())
			return
		}
		s.Effective.NoProxyAuto = *req.NoProxyAuto
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) runOllama(args []string) (string, int, error) {
	env, _, _ := config.BuildChildEnv(config.ChildEnvOptions{Existing: s.BaseEnv, Effective: s.Effective})
	var b strings.Builder
	code, err := ollamarunner.Run(context.Background(), ollamarunner.Options{
		Mode:        s.Effective.Mode,
		Host:        s.Effective.Host,
		OllamaExe:   s.Effective.OllamaExe,
		NoProxyAuto: s.Effective.NoProxyAuto,
		Unsafe:      s.Effective.Unsafe,
		Env:         env,
		Args:        args,
		Stdout:      &b,
		Stderr:      &b,
		Stdin:       strings.NewReader(""),
		Translator:  s.Translator,
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
