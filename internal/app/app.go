package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"cli_ollama_server/internal/config"
	"cli_ollama_server/internal/execollama"
	"cli_ollama_server/internal/i18n"
)

type globalOpts struct {
	Host      string
	Lang      string
	OllamaExe string
	Config    string
	Help      bool
	Version   bool
}

type argError struct {
	Kind string
	Flag string
}

func (e *argError) Error() string {
	if e == nil {
		return ""
	}
	return e.Kind + ": " + e.Flag
}

var (
	version = "dev"
	commit  = ""
)

func Run(args []string) int {
	opts, rest, err := parseGlobal(args)
	if err != nil {
		tr := i18n.New(i18n.DetectPreferredLang(""))
		if ae := (*argError)(nil); errors.As(err, &ae) {
			switch ae.Kind {
			case "unknown_flag":
				fmt.Fprintln(os.Stderr, tr.Sprintf("error.arg.unknown_flag", "flag", ae.Flag))
			case "missing_value":
				fmt.Fprintln(os.Stderr, tr.Sprintf("error.arg.missing_value", "flag", ae.Flag))
			default:
				fmt.Fprintln(os.Stderr, tr.Sprintf("error.invalid_args", "error", err.Error()))
			}
		} else {
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.invalid_args", "error", err.Error()))
		}
		fmt.Fprintln(os.Stderr, tr.Sprintf("help.try_help", "app", "ollama-remote"))
		return 2
	}

	cfg, cfgMeta, cfgErr := config.Load(config.LoadOptions{ExplicitConfigPath: opts.Config})

	lang := opts.Lang
	if lang == "" {
		if v := os.Getenv("OLLAMA_REMOTE_LANG"); v != "" {
			lang = v
		}
	}
	if lang == "" && cfg.Lang != "" {
		lang = cfg.Lang
	}
	lang = i18n.DetectPreferredLang(lang)
	tr := i18n.New(lang)

	if cfgErr != nil {
		if !(opts.Help || opts.Version || (len(rest) > 0 && rest[0] == "help")) {
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.config_load", "path", cfgMeta.PrimaryPath, "error", cfgErr.Error()))
			return 2
		}
	}

	if opts.Version {
		if commit != "" {
			fmt.Println(tr.Sprintf("app.version_commit", "version", version, "commit", commit))
		} else {
			fmt.Println(tr.Sprintf("app.version", "version", version))
		}
		return 0
	}

	if opts.Help || len(rest) == 0 || rest[0] == "help" {
		printHelp(tr)
		return 0
	}

	switch rest[0] {
	case "config":
		return runConfig(tr, cfg, cfgMeta, rest[1:])
	case "doctor":
		return runDoctor(tr, cfg, cfgMeta, opts)
	case "ui":
		return runUI(tr, cfg, cfgMeta, opts, rest[1:])
	}

	eff, effMeta := config.ResolveEffective(config.EffectiveOptions{
		GlobalHostFlag:      opts.Host,
		GlobalLangFlag:      opts.Lang,
		GlobalOllamaExeFlag: opts.OllamaExe,
		LoadedConfig:        cfg,
	})

	env, envMeta, envErr := config.BuildChildEnv(config.ChildEnvOptions{
		Existing:      os.Environ(),
		Effective:     eff,
		LoadedMeta:    cfgMeta,
		EffectiveMeta: effMeta,
	})
	if envErr != nil {
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.env_build", "error", envErr.Error()))
		return 2
	}

	code, runErr := execollama.Run(execollama.RunOptions{
		Args:      rest,
		Env:       env,
		OllamaExe: eff.OllamaExe,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		Stdin:     os.Stdin,
	})
	if runErr != nil {
		if errors.Is(runErr, execollama.ErrNotFound) {
			fmt.Fprintln(os.Stderr, tr.Sprintf("error.ollama_not_found", "hint", envMeta.OllamaHint()))
			return 127
		}
		if code != 0 {
			return code
		}
		fmt.Fprintln(os.Stderr, tr.Sprintf("error.ollama_failed", "error", runErr.Error()))
		return 1
	}
	return code
}

func parseGlobal(args []string) (globalOpts, []string, error) {
	var out globalOpts
	rest := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			rest = append(rest, args[i+1:]...)
			return out, rest, nil
		}
		if !strings.HasPrefix(a, "-") {
			rest = append(rest, args[i:]...)
			return out, rest, nil
		}

		switch {
		case a == "-h" || a == "--help":
			out.Help = true
		case a == "--version":
			out.Version = true
		case a == "--host":
			if i+1 >= len(args) {
				return out, nil, &argError{Kind: "missing_value", Flag: "--host"}
			}
			out.Host = args[i+1]
			i++
		case strings.HasPrefix(a, "--host="):
			out.Host = strings.TrimPrefix(a, "--host=")
		case a == "--lang":
			if i+1 >= len(args) {
				return out, nil, &argError{Kind: "missing_value", Flag: "--lang"}
			}
			out.Lang = args[i+1]
			i++
		case strings.HasPrefix(a, "--lang="):
			out.Lang = strings.TrimPrefix(a, "--lang=")
		case a == "--ollama-exe":
			if i+1 >= len(args) {
				return out, nil, &argError{Kind: "missing_value", Flag: "--ollama-exe"}
			}
			out.OllamaExe = args[i+1]
			i++
		case strings.HasPrefix(a, "--ollama-exe="):
			out.OllamaExe = strings.TrimPrefix(a, "--ollama-exe=")
		case a == "--config":
			if i+1 >= len(args) {
				return out, nil, &argError{Kind: "missing_value", Flag: "--config"}
			}
			out.Config = args[i+1]
			i++
		case strings.HasPrefix(a, "--config="):
			out.Config = strings.TrimPrefix(a, "--config=")
		default:
			return out, nil, &argError{Kind: "unknown_flag", Flag: a}
		}
	}
	return out, rest, nil
}

func printHelp(tr *i18n.Bundle) {
	fmt.Println(tr.Sprintf("help.usage", "app", "ollama-remote"))
	fmt.Println()
	fmt.Println(tr.Sprintf("help.what_is"))
	fmt.Println()
	fmt.Println(tr.Sprintf("help.global_flags"))
	fmt.Println(tr.Sprintf("help.flag.host"))
	fmt.Println(tr.Sprintf("help.flag.lang"))
	fmt.Println(tr.Sprintf("help.flag.ollama_exe"))
	fmt.Println(tr.Sprintf("help.flag.config"))
	fmt.Println(tr.Sprintf("help.flag.help"))
	fmt.Println(tr.Sprintf("help.flag.version"))
	fmt.Println()
	fmt.Println(tr.Sprintf("help.wrapper_cmds"))
	fmt.Println(tr.Sprintf("help.cmd.config"))
	fmt.Println(tr.Sprintf("help.cmd.doctor"))
	fmt.Println(tr.Sprintf("help.cmd.ui"))
	fmt.Println()
	fmt.Println(tr.Sprintf("help.examples"))
	fmt.Println(tr.Sprintf("help.example.list"))
	fmt.Println(tr.Sprintf("help.example.run"))
	fmt.Println(tr.Sprintf("help.example.host"))
	fmt.Println(tr.Sprintf("help.example.lang"))
	fmt.Println(tr.Sprintf("help.example.ui"))
}
