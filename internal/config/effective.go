package config

import (
	"net/url"
	"os"
	"strings"
)

type Effective struct {
	Host        string
	Lang        string
	OllamaExe   string
	NoProxyAuto bool
}

type EffectiveMeta struct {
	HostSource      string
	LangSource      string
	OllamaExeSource string
}

type EffectiveOptions struct {
	GlobalHostFlag      string
	GlobalLangFlag      string
	GlobalOllamaExeFlag string
	LoadedConfig        Config
}

func ResolveEffective(opts EffectiveOptions) (Effective, EffectiveMeta) {
	var out Effective
	var meta EffectiveMeta

	if strings.TrimSpace(opts.GlobalHostFlag) != "" {
		out.Host = strings.TrimSpace(opts.GlobalHostFlag)
		meta.HostSource = "flag"
	} else if v := strings.TrimSpace(os.Getenv("OLLAMA_HOST")); v != "" {
		out.Host = v
		meta.HostSource = "env"
	} else if strings.TrimSpace(opts.LoadedConfig.Host) != "" {
		out.Host = strings.TrimSpace(opts.LoadedConfig.Host)
		meta.HostSource = "config"
	} else {
		out.Host = "http://127.0.0.1:11434"
		meta.HostSource = "default"
	}

	if strings.TrimSpace(opts.GlobalLangFlag) != "" {
		out.Lang = strings.TrimSpace(opts.GlobalLangFlag)
		meta.LangSource = "flag"
	} else if v := strings.TrimSpace(os.Getenv("OLLAMA_REMOTE_LANG")); v != "" {
		out.Lang = v
		meta.LangSource = "env"
	} else if strings.TrimSpace(opts.LoadedConfig.Lang) != "" {
		out.Lang = strings.TrimSpace(opts.LoadedConfig.Lang)
		meta.LangSource = "config"
	} else {
		out.Lang = ""
		meta.LangSource = "auto"
	}

	if strings.TrimSpace(opts.GlobalOllamaExeFlag) != "" {
		out.OllamaExe = strings.TrimSpace(opts.GlobalOllamaExeFlag)
		meta.OllamaExeSource = "flag"
	} else if v := strings.TrimSpace(os.Getenv("OLLAMA_EXE")); v != "" {
		out.OllamaExe = v
		meta.OllamaExeSource = "env"
	} else if strings.TrimSpace(opts.LoadedConfig.OllamaExe) != "" {
		out.OllamaExe = strings.TrimSpace(opts.LoadedConfig.OllamaExe)
		meta.OllamaExeSource = "config"
	} else {
		out.OllamaExe = ""
		meta.OllamaExeSource = "auto"
	}

	if opts.LoadedConfig.NoProxyAuto != nil {
		out.NoProxyAuto = *opts.LoadedConfig.NoProxyAuto
	}
	return out, meta
}

type ChildEnvOptions struct {
	Existing      []string
	Effective     Effective
	LoadedMeta    LoadMeta
	EffectiveMeta EffectiveMeta
}

type ChildEnvMeta struct {
	HostURL      *url.URL
	OllamaExeSet bool
}

func (m ChildEnvMeta) OllamaHint() string {
	if m.OllamaExeSet {
		return ""
	}
	return "(hint: install Ollama and ensure `ollama` is on PATH, or set OLLAMA_EXE)"
}

func BuildChildEnv(opts ChildEnvOptions) ([]string, ChildEnvMeta, error) {
	env := make(map[string]string, len(opts.Existing))
	for _, kv := range opts.Existing {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		env[k] = v
	}

	meta := ChildEnvMeta{OllamaExeSet: strings.TrimSpace(opts.Effective.OllamaExe) != ""}

	if strings.TrimSpace(opts.Effective.Host) != "" {
		env["OLLAMA_HOST"] = strings.TrimSpace(opts.Effective.Host)
		if u, err := url.Parse(env["OLLAMA_HOST"]); err == nil {
			meta.HostURL = u
			if opts.Effective.NoProxyAuto {
				addNoProxy(env, u)
			}
		}
	}
	if meta.OllamaExeSet {
		env["OLLAMA_EXE"] = strings.TrimSpace(opts.Effective.OllamaExe)
	}

	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out, meta, nil
}

func addNoProxy(env map[string]string, u *url.URL) {
	host := u.Hostname()
	if host == "" {
		return
	}
	cur := env["NO_PROXY"]
	if cur == "" {
		env["NO_PROXY"] = host
		return
	}
	parts := strings.Split(cur, ",")
	for _, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), host) {
			return
		}
	}
	env["NO_PROXY"] = cur + "," + host
}
