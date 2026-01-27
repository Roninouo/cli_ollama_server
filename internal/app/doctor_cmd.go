package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"cli_ollama_server/internal/config"
	"cli_ollama_server/internal/execollama"
	"cli_ollama_server/internal/i18n"
	"cli_ollama_server/internal/ollamaapi"
)

func runDoctor(tr *i18n.Bundle, loaded config.Config, meta config.LoadMeta, opts globalOpts) int {
	eff, _ := config.ResolveEffective(config.EffectiveOptions{
		GlobalHostFlag:      opts.Host,
		GlobalLangFlag:      opts.Lang,
		GlobalOllamaExeFlag: opts.OllamaExe,
		GlobalModeFlag:      opts.Mode,
		GlobalUnsafeFlag:    opts.Unsafe,
		LoadedConfig:        loaded,
	})
	if m, merr := config.NormalizeMode(eff.Mode); merr != nil {
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.invalid_mode", "mode", eff.Mode))
		return 2
	} else {
		eff.Mode = m
	}
	if _, herr := config.ParseHostURL(eff.Host); herr != nil {
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.invalid_host", "host", eff.Host, "error", herr.Error()))
		return 2
	}

	env, _, envErr := config.BuildChildEnv(config.ChildEnvOptions{Existing: os.Environ(), Effective: eff, LoadedMeta: meta})
	if envErr != nil {
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.env_build", "error", envErr.Error()))
		return 2
	}

	fmt.Println(tr.Sprintf("doctor.host", "value", eff.Host))
	fmt.Println(tr.Sprintf("doctor.lang", "value", tr.Lang()))
	fmt.Println(tr.Sprintf("doctor.mode", "value", eff.Mode))
	fmt.Println(tr.Sprintf("doctor.unsafe", "value", fmtBool(eff.Unsafe)))

	selected := eff.Mode
	if selected == "auto" {
		if _, err := execollama.ResolveExecutable(eff.OllamaExe); err == nil {
			selected = "wrapper"
		} else {
			selected = "native"
		}
	}
	fmt.Println(tr.Sprintf("doctor.selected_mode", "value", selected))

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	apiOK := false
	if u, _ := config.ParseHostURL(eff.Host); u != nil {
		v, verr := ollamaapi.NewClient(u, eff.NoProxyAuto).Version(ctx)
		if verr != nil {
			fmt.Fprintln(os.Stderr, tr.Sprintf("doctor.api_version_failed", "error", verr.Error()))
		} else {
			apiOK = true
			fmt.Println(tr.Sprintf("doctor.api_version", "value", v))
		}
	}

	// Wrapper CLI detection (works in both modes).
	exe, xerr := execollama.ResolveExecutable(eff.OllamaExe)
	if xerr != nil {
		fmt.Fprintln(os.Stderr, tr.Sprintf("doctor.ollama_cli", "value", tr.Sprintf("doctor.value.not_found")))
		if selected == "wrapper" {
			return 127
		}
		if !apiOK {
			return 1
		}
		return 0
	}
	fmt.Println(tr.Sprintf("doctor.ollama_cli", "value", exe))

	// Wrapper CLI smoke test (local only).
	code, err := execollama.Run(execollama.RunOptions{
		Context:   ctx,
		Args:      []string{"--version"},
		Env:       env,
		OllamaExe: exe,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		Stdin:     os.Stdin,
	})
	if err != nil {
		if errors.Is(err, execollama.ErrNotFound) {
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.ollama_not_found", "hint", ""))
			return 127
		}
		fmt.Fprintln(os.Stderr, tr.Sprintf("doctor.ollama_failed", "error", err.Error()))
		if code != 0 {
			return code
		}
		return 1
	}
	if !apiOK {
		return 1
	}
	return 0
}
