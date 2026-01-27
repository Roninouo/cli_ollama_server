package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

type Config struct {
	Host        string `toml:"host"`
	Lang        string `toml:"lang"`
	OllamaExe   string `toml:"ollama_exe"`
	Mode        string `toml:"mode"`
	NoProxyAuto *bool  `toml:"no_proxy_auto"`
	Unsafe      *bool  `toml:"unsafe"`
}

type LoadOptions struct {
	ExplicitConfigPath string
}

type LoadMeta struct {
	PrimaryPath string
	UsedFiles   []string
}

func Load(opts LoadOptions) (Config, LoadMeta, error) {
	meta := LoadMeta{PrimaryPath: DefaultUserConfigPath()}
	if opts.ExplicitConfigPath != "" {
		meta.PrimaryPath = opts.ExplicitConfigPath
	}

	// If user explicitly selected a config file, only load that one.
	if opts.ExplicitConfigPath != "" {
		c, err := readTomlIfExists(opts.ExplicitConfigPath)
		if err != nil {
			return Config{}, meta, err
		}
		if fileExists(opts.ExplicitConfigPath) {
			meta.UsedFiles = append(meta.UsedFiles, opts.ExplicitConfigPath)
		}
		return c, meta, nil
	}

	// Lowest precedence: user config.
	userCfg, err := readTomlIfExists(meta.PrimaryPath)
	if err != nil {
		return Config{}, meta, err
	}
	if fileExists(meta.PrimaryPath) {
		meta.UsedFiles = append(meta.UsedFiles, meta.PrimaryPath)
	}
	out := userCfg

	// Project config overrides user config.
	cwd, err := os.Getwd()
	if err == nil {
		projToml := filepath.Join(cwd, ".ollama-remote.toml")
		projEnvDefault := filepath.Join(cwd, ".env")
		projEnvTool := filepath.Join(cwd, ".ollama-remote.env")

		// Project env files (higher overrides lower): .env then .ollama-remote.env
		if envm, eerr := readEnvIfExists(projEnvDefault); eerr != nil {
			return Config{}, meta, eerr
		} else if len(envm) > 0 {
			meta.UsedFiles = append(meta.UsedFiles, projEnvDefault)
			out = mergeEnvIntoConfig(out, envm)
		}

		if envm, eerr := readEnvIfExists(projEnvTool); eerr != nil {
			return Config{}, meta, eerr
		} else if len(envm) > 0 {
			meta.UsedFiles = append(meta.UsedFiles, projEnvTool)
			out = mergeEnvIntoConfig(out, envm)
		}

		if projCfg, perr := readTomlIfExists(projToml); perr != nil {
			return Config{}, meta, perr
		} else {
			if fileExists(projToml) {
				meta.UsedFiles = append(meta.UsedFiles, projToml)
			}
			out = mergeConfig(out, projCfg)
		}
	}

	return out, meta, nil
}

func DefaultUserConfigPath() string {
	d, err := os.UserConfigDir()
	if err != nil || d == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			return "config.toml"
		}
		return filepath.Join(home, ".config", "ollama-remote", "config.toml")
	}
	return filepath.Join(d, "ollama-remote", "config.toml")
}

func readTomlIfExists(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var c Config
	if err := toml.Unmarshal(b, &c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return c, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func mergeConfig(base, override Config) Config {
	if strings.TrimSpace(override.Host) != "" {
		base.Host = override.Host
	}
	if strings.TrimSpace(override.Lang) != "" {
		base.Lang = override.Lang
	}
	if strings.TrimSpace(override.OllamaExe) != "" {
		base.OllamaExe = override.OllamaExe
	}
	if strings.TrimSpace(override.Mode) != "" {
		base.Mode = override.Mode
	}
	if override.NoProxyAuto != nil {
		base.NoProxyAuto = override.NoProxyAuto
	}
	if override.Unsafe != nil {
		base.Unsafe = override.Unsafe
	}
	return base
}
