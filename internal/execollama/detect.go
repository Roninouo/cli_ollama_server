package execollama

import (
	"errors"
	"os/exec"
	"strings"
)

// ResolveExecutable resolves the ollama executable path.
//
// If ollamaExe is empty, it resolves "ollama" from PATH.
// If ollamaExe is set, it is treated as a command/path and resolved via LookPath.
func ResolveExecutable(ollamaExe string) (string, error) {
	name := strings.TrimSpace(ollamaExe)
	if name == "" {
		name = "ollama"
	}
	p, err := exec.LookPath(name)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return p, nil
}
