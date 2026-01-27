# Commands

`ollama-remote` supports a hybrid execution model:

1) Wrapper mode: uses the official `ollama` CLI when available
2) Native mode: falls back to the Ollama REST API when the `ollama` CLI is not available

See: `docs/hybrid.md`

## Wrapper commands

### `config`

- `ollama-remote config show`
- `ollama-remote config init`
- `ollama-remote config set <host|lang|ollama_exe|no_proxy_auto> <value>`
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

Notes:

- In wrapper mode, unknown commands/flags are forwarded to `ollama`.
- In native mode, only a documented subset of commands is supported (unsupported commands fail fast).
