package execollama

import (
	"context"
	"errors"
	"io"
	"os/exec"
)

var ErrNotFound = errors.New("ollama not found")

type RunOptions struct {
	Context   context.Context
	Args      []string
	Env       []string
	OllamaExe string
	Stdout    io.Writer
	Stderr    io.Writer
	Stdin     io.Reader
}

func Run(opts RunOptions) (int, error) {
	exe := opts.OllamaExe
	if exe == "" {
		p, err := exec.LookPath("ollama")
		if err != nil {
			return 0, ErrNotFound
		}
		exe = p
	}

	if len(opts.Args) == 0 {
		return 0, nil
	}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, exe, opts.Args...)
	cmd.Env = opts.Env
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr
	cmd.Stdin = opts.Stdin

	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), err
	}
	return 0, err
}
