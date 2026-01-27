# cli_ollama_server (ollama-remote)

`ollama-remote` is a small, production-friendly CLI that targets a remote Ollama server.

It uses a hybrid execution model:

- Wrapper mode (preferred): uses the official `ollama` CLI when available
- Native mode (fallback): talks directly to the Ollama REST API when `ollama` is not available

It resolves configuration (`host`, `lang`, `ollama_exe`, `mode`, `unsafe`) and then either runs the upstream CLI or calls the REST API.

## What this tool does

- In wrapper mode, runs the upstream `ollama` CLI with a resolved `OLLAMA_HOST`
- Falls back to a minimal built-in REST client for common commands when `ollama` is not present
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

If theres any issue related with permissions consider and trust issues with PowerShell try: `.\ollama-remote`, go to "get-help about_Command_Precedence" to obtain more detailed info.

3) Initialize config (optional) and set your host (defaults to `http://127.0.0.1:11434`):

```powershell
ollama-remote config init
ollama-remote config set host https://ollama.example.com:11434
```

4) Use it:

```powershell
ollama-remote list
ollama-remote run llama3:8b
ollama-remote --host https://ollama.example.com:11434 ps
```

## Quick Start (macOS / Linux)

1) Download the latest release zip and extract it.
2) Add the folder to your PATH:

```sh
chmod +x ./install.sh ./ollama-remote
./install.sh
```

3) Set your host (optional, defaults to `http://127.0.0.1:11434`):

```sh
ollama-remote config init
ollama-remote config set host https://ollama.example.com:11434
```

4) Use it:

```sh
ollama-remote list
ollama-remote run llama3:8b
ollama-remote --host https://ollama.example.com:11434 ps
```

## Execution modes (hybrid)

`ollama-remote` can execute in three modes:

- `--mode auto` (default): use wrapper if `ollama` is available, otherwise native
- `--mode wrapper`: force the upstream `ollama` CLI (full upstream feature set)
- `--mode native`: never uses `ollama`; supports a documented subset via REST

Examples:

```bash
# Auto (default): wrapper if available, else native
ollama-remote list

# Force native REST client
ollama-remote --mode native list

# Force wrapper mode (requires local ollama CLI)
ollama-remote --mode wrapper list
```

See `docs/hybrid.md` for the native support matrix.

## Command overview

- `ollama-remote config init|show|set|path`: manage configuration
- `ollama-remote doctor`: validate host/mode and check local CLI + remote API
- `ollama-remote ui`: start the optional localhost web UI
- Any other args (wrapper mode): forwarded to the upstream `ollama` CLI

## Using the local Ollama CLI (Windows: ollama.exe)

Wrapper mode requires the upstream Ollama CLI to be installed locally.

1) Verify you have it:

```powershell
where.exe ollama
ollama --version
ollama-remote doctor
```

2) If `ollama` is not on PATH, point `ollama-remote` to `ollama.exe`:

```powershell
# One-off
ollama-remote --mode wrapper --ollama-exe "C:\Program Files\Ollama\ollama.exe" list

# Or via environment (current shell)
$env:OLLAMA_EXE = "C:\Program Files\Ollama\ollama.exe"
ollama-remote --mode wrapper list

# Persist in user config
ollama-remote config set ollama_exe "C:\Program Files\Ollama\ollama.exe"
ollama-remote config set mode wrapper
```

TOML tip (Windows paths): prefer single quotes to avoid backslash escaping:

```toml
ollama_exe = 'C:\Program Files\Ollama\ollama.exe'
```

3) You can also use the upstream CLI directly (without `ollama-remote`) by setting `OLLAMA_HOST`:

```powershell
$env:OLLAMA_HOST = "https://ollama.example.com:11434"
& "C:\Program Files\Ollama\ollama.exe" list
```

## Documentation

- `docs/installation.md`
- `docs/configuration.md`
- `docs/commands.md`
- `docs/hybrid.md`
- `docs/i18n.md`
- `docs/ui.md`
- `docs/troubleshooting.md`

## Legacy wrappers

If you still need the original script-based wrappers, see:

- `legacy/ollama-remote.bat`
- `legacy/ollama-remote.ps1`
