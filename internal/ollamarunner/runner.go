package ollamarunner

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"cli_ollama_server/internal/config"
	"cli_ollama_server/internal/execollama"
	"cli_ollama_server/internal/i18n"
	"cli_ollama_server/internal/ollamaapi"
)

type Options struct {
	Mode        string
	Host        string
	OllamaExe   string
	NoProxyAuto bool
	Unsafe      bool

	Env        []string
	Args       []string
	Stdout     io.Writer
	Stderr     io.Writer
	Stdin      io.Reader
	Translator *i18n.Bundle
}

func Run(ctx context.Context, opts Options) (int, error) {
	mode := strings.TrimSpace(opts.Mode)
	if mode == "" {
		mode = "auto"
	}

	if mode == "auto" {
		if _, err := execollama.ResolveExecutable(opts.OllamaExe); err == nil {
			mode = "wrapper"
		} else {
			mode = "native"
		}
	}

	switch mode {
	case "wrapper":
		exe, err := execollama.ResolveExecutable(opts.OllamaExe)
		if err != nil {
			return 0, err
		}
		if ctx == nil {
			ctx = context.Background()
		}
		return execollama.Run(execollama.RunOptions{
			Context:   ctx,
			Args:      opts.Args,
			Env:       opts.Env,
			OllamaExe: exe,
			Stdout:    opts.Stdout,
			Stderr:    opts.Stderr,
			Stdin:     opts.Stdin,
		})
	case "native":
		return runNative(ctx, opts)
	default:
		return 2, fmt.Errorf("invalid mode: %s", mode)
	}
}

func runNative(ctx context.Context, opts Options) (int, error) {
	if len(opts.Args) == 0 {
		return 0, nil
	}
	tr := opts.Translator
	if tr == nil {
		tr = i18n.New("en")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	baseURL, err := config.ParseHostURL(opts.Host)
	if err != nil {
		return 2, err
	}
	client := ollamaapi.NewClient(baseURL, opts.NoProxyAuto)

	cmd := opts.Args[0]
	switch cmd {
	case "--version":
		v, err := client.Version(ctx)
		if err != nil {
			return 1, err
		}
		fmt.Fprintln(opts.Stdout, v)
		return 0, nil
	case "list":
		models, err := client.Tags(ctx)
		if err != nil {
			return 1, err
		}
		fmt.Fprint(opts.Stdout, ollamaapi.FormatTags(models))
		return 0, nil
	case "ps":
		procs, err := client.PS(ctx)
		if err != nil {
			return 1, err
		}
		fmt.Fprint(opts.Stdout, ollamaapi.FormatPS(procs))
		return 0, nil
	case "show":
		if len(opts.Args) < 2 {
			return 2, errors.New(tr.Sprintf("error.native.usage_show"))
		}
		resp, err := client.Show(ctx, strings.TrimSpace(opts.Args[1]))
		if err != nil {
			return 1, err
		}
		b, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Fprintln(opts.Stdout, string(b))
		return 0, nil
	case "pull":
		if len(opts.Args) < 2 {
			return 2, errors.New(tr.Sprintf("error.native.usage_pull"))
		}
		if !opts.Unsafe {
			return 2, errors.New(tr.Sprintf("error.native.pull_requires_unsafe"))
		}
		if err := client.Pull(ctx, strings.TrimSpace(opts.Args[1]), opts.Stdout); err != nil {
			return 1, err
		}
		return 0, nil
	case "run":
		if len(opts.Args) < 2 {
			return 2, errors.New(tr.Sprintf("error.native.usage_run"))
		}
		model := strings.TrimSpace(opts.Args[1])
		rest := opts.Args[2:]
		if len(rest) > 0 && rest[0] == "--" {
			rest = rest[1:]
		}
		prompt := strings.TrimSpace(strings.Join(rest, " "))
		if prompt == "" {
			if stdin, rerr := readStdinIfPiped(opts.Stdin); rerr != nil {
				return 1, rerr
			} else if strings.TrimSpace(stdin) != "" {
				prompt = stdin
			}
		}
		if prompt == "" {
			return 2, errors.New(tr.Sprintf("error.native.run_requires_prompt"))
		}
		req := ollamaapi.GenerateRequest{Model: model, Prompt: prompt, Stream: true}
		if err := client.Generate(ctx, req, opts.Stdout); err != nil {
			return 1, err
		}
		return 0, nil
	default:
		return 2, fmt.Errorf(tr.Sprintf("error.native.unsupported", "cmd", cmd))
	}
}

func readStdinIfPiped(r io.Reader) (string, error) {
	if r == nil {
		return "", nil
	}
	f, ok := r.(*os.File)
	if ok {
		if st, err := f.Stat(); err == nil {
			if (st.Mode() & os.ModeCharDevice) != 0 {
				return "", nil
			}
		}
	}
	br := bufio.NewReader(r)
	b, err := io.ReadAll(br)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
