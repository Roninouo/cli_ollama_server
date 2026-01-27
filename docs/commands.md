# Commands

`ollama-remote` supports a hybrid execution model:

1) Wrapper mode: uses the official `ollama` CLI when available
2) Native mode: falls back to the Ollama REST API when the `ollama` CLI is not available

See: `docs/hybrid.md`

## Wrapper commands

### `config`

- `ollama-remote config show`
- `ollama-remote config init`
- `ollama-remote config set <host|lang|ollama_exe|mode|no_proxy_auto|unsafe> <value>`
- `ollama-remote config path`

### `doctor`

- `ollama-remote doctor`

Checks basic setup and runs `ollama --version` using the resolved configuration.

### `ui`

- `ollama-remote ui`

Starts an optional local web UI bound to `127.0.0.1` and opens your browser.

## Passthrough examples

```bash
ollama-remote list
ollama-remote ps
ollama-remote run llama3:8b
ollama-remote --host https://ollama.example.com:11434 pull llama3:8b
ollama-remote --mode native list
ollama-remote --mode native --unsafe pull llama3:8b
```

## Using the local Ollama CLI (wrapper mode)

Wrapper mode uses your local `ollama` / `ollama.exe` binary and forwards args to it (so you get the full upstream feature set).

```powershell
# If ollama.exe is on PATH
ollama-remote --mode wrapper list

# If ollama.exe is NOT on PATH
ollama-remote --mode wrapper --ollama-exe "C:\Program Files\Ollama\ollama.exe" list
```

Tip: you can also run the upstream CLI directly by setting `OLLAMA_HOST`:

```powershell
$env:OLLAMA_HOST = "https://ollama.example.com:11434"
ollama list
```

## Native mode (REST) supported subset

Native mode does not require a local Ollama installation, but it only supports a subset:

- `--version`
- `list`
- `ps`
- `show <model>`
- `run <model> [--] <prompt>` (prompt arg or piped stdin; no interactive session)
- `pull <model>` only with `--unsafe`

Notes:

- In wrapper mode, unknown commands/flags are forwarded to `ollama`.
- In native mode, only a documented subset of commands is supported (unsupported commands fail fast).
