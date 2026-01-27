# Configuration

## What is configurable

- `host`: the Ollama server address (mapped to `OLLAMA_HOST` for the spawned `ollama` process)
- `lang`: UI/help/error language for this tool (`en`, `es`, `de`, or `auto`)
- `ollama_exe`: full path to the official Ollama CLI executable
- `mode`: execution mode (`auto`, `wrapper`, `native`)
- `no_proxy_auto`: if `true`, adds the host's hostname to `NO_PROXY` for the spawned process only
- `unsafe`: if `true`, enables mutating/advanced operations in native mode (disabled by default)

## Precedence (highest to lowest)

1) CLI flags: `--host`, `--lang`, `--ollama-exe`, `--mode`, `--unsafe`, `--config`
2) Environment: `OLLAMA_HOST`, `OLLAMA_EXE`, `OLLAMA_REMOTE_LANG`, `OLLAMA_REMOTE_MODE`, `OLLAMA_REMOTE_UNSAFE`
3) Project files in the current directory:

- `.env` (optional)
- `.ollama-remote.env`
- `.ollama-remote.toml`

4) User config file (default):

- Windows: `%APPDATA%\\ollama-remote\\config.toml`
- Linux: `~/.config/ollama-remote/config.toml`
- macOS: `~/Library/Application Support/ollama-remote/config.toml`

5) Safe defaults:

- `host = http://127.0.0.1:11434`
- `lang = auto` (fallback `en`)
- `mode = auto`
- `unsafe = false`

## Examples

Create a user config file:

```bash
ollama-remote config init
```

Set the host:

```bash
ollama-remote config set host https://ollama.example.com:11434
```

Project-local `.ollama-remote.env`:

```env
OLLAMA_HOST=https://ollama.example.com:11434
OLLAMA_REMOTE_LANG=de
OLLAMA_REMOTE_MODE=auto
# OLLAMA_REMOTE_UNSAFE=1
```

Mode notes:

- `mode=auto` prefers wrapper mode if `ollama` is available, otherwise native mode.
- If you set `OLLAMA_EXE` / `ollama_exe` and it cannot be resolved (and `mode != native`), `ollama-remote` fails fast instead of silently switching modes.
