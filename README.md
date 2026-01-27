# cli_ollama_server (ollama-remote)

`ollama-remote` is a small, production-friendly wrapper around the official `ollama` CLI.

It resolves configuration (`host`, `lang`, `ollama_exe`) and runs `ollama` with your arguments.

## What this tool does

- Runs the upstream `ollama` CLI with a resolved `OLLAMA_HOST`
- Provides config management, basic diagnostics, and an optional local web UI
- Propagates exit codes (CI-friendly)

## What this tool does NOT do

- It does not run or manage an Ollama server
- It does not reimplement the Ollama API
- It does not ship Ollama itself

## Quick Start (Windows)

1) Download the latest release zip and extract it.
2) Add the folder to your User PATH:

```powershell
pwsh -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
```

## Quick Start (macOS / Linux)

1) Download the latest release zip and extract it.
2) Add the folder to your PATH:

```sh
chmod +x ./install.sh ./ollama-remote
./install.sh
```

3) Set your host (optional, defaults to `http://127.0.0.1:11434`):

```bash
ollama-remote config init
ollama-remote config set host https://ollama.example.com:11434
```

4) Use it:

```bash
ollama-remote list
ollama-remote run llama3:8b
ollama-remote --host https://ollama.example.com:11434 ps
```

## Documentation

- `docs/installation.md`
- `docs/configuration.md`
- `docs/commands.md`
- `docs/i18n.md`
- `docs/ui.md`
- `docs/troubleshooting.md`

## Legacy wrappers

If you still need the original script-based wrappers, see:

- `legacy/ollama-remote.bat`
- `legacy/ollama-remote.ps1`
