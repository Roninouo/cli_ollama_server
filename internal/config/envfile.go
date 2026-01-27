package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

func readEnvIfExists(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	out := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" {
			continue
		}
		v = strings.Trim(v, "\"'")
		out[k] = v
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func mergeEnvIntoConfig(base Config, env map[string]string) Config {
	if v := strings.TrimSpace(env["OLLAMA_HOST"]); v != "" {
		base.Host = v
	}
	if v := strings.TrimSpace(env["OLLAMA_EXE"]); v != "" {
		base.OllamaExe = v
	}
	if v := strings.TrimSpace(env["OLLAMA_REMOTE_LANG"]); v != "" {
		base.Lang = v
	}
	if v := strings.TrimSpace(env["OLLAMA_REMOTE_NO_PROXY_AUTO"]); v != "" {
		b := (v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes"))
		base.NoProxyAuto = &b
	}
	return base
}
