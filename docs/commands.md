# Commands

`ollama-remote` supports two modes:

1) Wrapper commands implemented by this tool
2) Passthrough mode: any unknown command/args are forwarded to the official `ollama` CLI

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
```
