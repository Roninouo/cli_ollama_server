package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

func InitUserConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b0 := false
	c := Config{
		Host:        "http://127.0.0.1:11434",
		Lang:        "",
		OllamaExe:   "",
		NoProxyAuto: &b0,
	}
	b, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func SetUserConfig(path, key, val string) error {
	key = strings.TrimSpace(strings.ToLower(key))
	if key == "" {
		return fmt.Errorf("empty key")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	c, err := readTomlIfExists(path)
	if err != nil {
		return err
	}

	switch key {
	case "host":
		c.Host = val
	case "lang":
		c.Lang = val
	case "ollama_exe":
		c.OllamaExe = val
	case "no_proxy_auto":
		v := strings.TrimSpace(val)
		b := (v == "1" || strings.EqualFold(v, "true") || strings.EqualFold(v, "yes"))
		c.NoProxyAuto = &b
	default:
		return &UnknownKeyError{Key: key}
	}

	b, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
