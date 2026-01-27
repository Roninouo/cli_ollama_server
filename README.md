# cli_ollama_server

Thin, dependency-light wrappers around the official `ollama.exe` CLI to default to a remote Ollama server.

Default server:

- `http://10.65.117.238:11434`

## Quick Start

1) Add this folder to your User PATH:

```powershell
pwsh -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
```

2) Use the wrapper:

```bat
ollama-remote list
ollama-remote run llama3:8b
ollama-remote --host http://10.65.117.238:11434 ps
```

## Behavior

- If `OLLAMA_HOST` is already set, the wrapper will not override it.
- You can override per call with `--host <url>` or `--host=<url>`.
- Exit codes are propagated (CI-friendly).

## Environment Variables

- `OLLAMA_HOST`: target server URL (e.g. `http://10.65.117.238:11434`)
- `OLLAMA_EXE`: full path to `ollama.exe` (useful if `ollama` is not on PATH)

## Files

- `ollama-remote.bat`: primary Batch wrapper
- `ollama-remote.ps1`: optional PowerShell wrapper (more flexible arg parsing)
- `install.ps1`: add this folder to User PATH (no admin)
- `uninstall.ps1`: remove this folder from User PATH
