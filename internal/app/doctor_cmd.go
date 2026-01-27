package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"cli_ollama_server/internal/config"
	"cli_ollama_server/internal/execollama"
	"cli_ollama_server/internal/i18n"
)

func runDoctor(tr *i18n.Bundle, loaded config.Config, meta config.LoadMeta, opts globalOpts) int {
	eff, _ := config.ResolveEffective(config.EffectiveOptions{
		GlobalHostFlag:      opts.Host,
		GlobalLangFlag:      opts.Lang,
		GlobalOllamaExeFlag: opts.OllamaExe,
		LoadedConfig:        loaded,
	})

	env, _, envErr := config.BuildChildEnv(config.ChildEnvOptions{Existing: os.Environ(), Effective: eff, LoadedMeta: meta})
	if envErr != nil {
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.env_build", "error", envErr.Error()))
		return 2
	}

	fmt.Println(tr.Sprintf("doctor.host", "value", eff.Host))
	fmt.Println(tr.Sprintf("doctor.lang", "value", tr.Lang()))

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	code, err := execollama.Run(execollama.RunOptions{
		Context:   ctx,
		Args:      []string{"--version"},
		Env:       env,
		OllamaExe: eff.OllamaExe,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		Stdin:     os.Stdin,
	})
	if err != nil {
		if err == execollama.ErrNotFound {
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.ollama_not_found", "hint", ""))
			return 127
		}
		fmt.Fprintln(os.Stderr, tr.Sprintf("doctor.ollama_failed", "error", err.Error()))
		if code != 0 {
			return code
		}
		return 1
	}
	return 0
}
