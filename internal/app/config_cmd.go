package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"cli_ollama_server/internal/config"
	"cli_ollama_server/internal/i18n"
)

func runConfig(tr *i18n.Bundle, loaded config.Config, meta config.LoadMeta, args []string) int {
	sub := "show"
	if len(args) > 0 {
		sub = args[0]
		args = args[1:]
	}

	switch sub {
	case "path":
		fmt.Println(meta.PrimaryPath)
		return 0
	case "show":
		eff, _ := config.ResolveEffective(config.EffectiveOptions{LoadedConfig: loaded})
		fmt.Println(tr.Sprintf("config.path", "path", meta.PrimaryPath))
		fmt.Println(tr.Sprintf("config.host", "value", eff.Host))
		langVal := eff.Lang
		if strings.TrimSpace(langVal) == "" {
			langVal = tr.Sprintf("config.value.auto")
		}
		fmt.Println(tr.Sprintf("config.lang", "value", langVal))
		if eff.OllamaExe == "" {
			fmt.Println(tr.Sprintf("config.ollama_exe", "value", tr.Sprintf("config.value.auto")))
		} else {
			fmt.Println(tr.Sprintf("config.ollama_exe", "value", eff.OllamaExe))
		}
		modeVal := eff.Mode
		if strings.TrimSpace(modeVal) == "" {
			modeVal = "auto"
		}
		fmt.Println(tr.Sprintf("config.mode", "value", modeVal))
		fmt.Println(tr.Sprintf("config.no_proxy_auto", "value", fmtBool(eff.NoProxyAuto)))
		fmt.Println(tr.Sprintf("config.unsafe", "value", fmtBool(eff.Unsafe)))
		return 0
	case "init":
		if err := config.InitUserConfig(meta.PrimaryPath); err != nil {
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.config_init", "path", meta.PrimaryPath, "error", err.Error()))
			return 1
		}
		fmt.Println(tr.Sprintf("config.inited", "path", meta.PrimaryPath))
		return 0
	case "set":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.config_set_usage"))
			return 2
		}
		key, val := args[0], args[1]
		if err := config.SetUserConfig(meta.PrimaryPath, key, val); err != nil {
			var uke *config.UnknownKeyError
			if errors.As(err, &uke) {
				fmt.Fprintln(os.Stderr, tr.Sprintf("error.config_unknown_key", "key", uke.Key))
				return 2
			}
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.config_set", "error", err.Error()))
			return 1
		}
		fmt.Println(tr.Sprintf("config.set_ok", "key", key))
		return 0
	default:
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.unknown_subcommand", "sub", sub))
		return 2
	}
}

func fmtBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
