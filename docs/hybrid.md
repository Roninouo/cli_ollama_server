# Hybrid Execution Model

`ollama-remote` is a single CLI with one configuration and one UX. Internally it can execute commands in two ways.

1) Wrapper mode: spawn the official `ollama` CLI (preferred)
2) Native mode: call the Ollama REST API directly (fallback)

## Architecture Overview

```
                 +-----------------------+
User CLI args -> |  ollama-remote (UX)   |
                 |  - flags/env/.env/toml|
                 |  - validation         |
                 +-----------+-----------+
                             |
                             v
                   +-------------------+
                   | Mode Selector     |
                   | auto|wrapper|native
                   +----+----------+---+
                        |          |
          wrapper (exec)|          | native (HTTP)
                        v          v
                +-----------+   +----------------+
                | ollama    |   | REST client     |
                | (upstream)|   | /api/* endpoints|
                +-----+-----+   +--------+--------+
                      |                  |
                      v                  v
                 Remote Ollama Server (same host config)
```

Key properties:

- One config/UX surface: `--host`, `--mode`, `--unsafe`, etc.
- One security boundary: no shell invocation; JSON is encoded; validation is strict.
- Explicit parity: native mode is intentionally a subset of the upstream CLI.

## Capability Detection + Mode Selection

Mode input precedence (highest to lowest):

1) CLI flag: `--mode`
2) Environment: `OLLAMA_REMOTE_MODE`
3) Config: `mode = "..."`
4) Default: `auto`

Mode behavior:

- `--mode=wrapper`
  - Requires the official `ollama` executable to be resolvable
  - If missing: exit `127` and print an actionable error

- `--mode=native`
  - Never uses the `ollama` executable
  - Only a documented subset of commands is supported

- `--mode=auto` (default)
  - If `ollama` is available: use wrapper mode
  - Otherwise: use native mode

### Using the local Ollama CLI (ollama.exe)

Wrapper mode runs the upstream CLI on your machine (Windows: `ollama.exe`) and forwards the args.

Resolution rules:

- If `--ollama-exe` / `OLLAMA_EXE` / `ollama_exe` is set: that path must exist (unless `mode=native`).
- Otherwise `ollama-remote` looks for `ollama` on `PATH`.

Examples (Windows):

```powershell
# Force wrapper mode using PATH
ollama-remote --mode wrapper list

# Force wrapper mode using an explicit path
ollama-remote --mode wrapper --ollama-exe "C:\Program Files\Ollama\ollama.exe" list
```

Validation rules (no silent downgrades):

- If `OLLAMA_EXE` / `--ollama-exe` is set (and mode is not `native`) but cannot be resolved: fail fast (exit `2`).

## Command Support Matrix

Legend:

- Yes: implemented
- No: not implemented (requires wrapper mode)
- Gated: supported only with `--unsafe` (or `unsafe=true`)

| Command | Wrapper Mode | Native Mode | Notes |
|--------:|:------------:|:-----------:|-------|
| `list` | Yes | Yes | Native prints a simple table based on `/api/tags` |
| `ps` | Yes | Yes | Native prints a simple table based on `/api/ps` |
| `run <model> [prompt]` | Yes | Yes (non-interactive) | Native requires a prompt arg or piped stdin; interactive sessions require wrapper |
| `show <model>` | Yes | Yes | Native prints pretty JSON from `/api/show` |
| `pull <model>` | Yes | Gated | Native uses `/api/pull`; disabled by default |
| `*` (other ollama commands/flags) | Yes | No | Install `ollama` or use `--mode=wrapper` |

Wrapper-only commands implemented by this tool:

- `config ...`
- `doctor`
- `ui`

## Security Considerations (By Mode)

Common (all modes):

- No shell execution: wrapper mode uses `exec.CommandContext` with argv arrays.
- Configuration is validated (`host` must be an absolute http(s) URL).
- No implicit privilege escalation: advanced capabilities are opt-in.

Wrapper mode:

- Security posture tracks the upstream `ollama` CLI behavior.
- `OLLAMA_HOST` is injected via environment (scoped to the child process).
- The tool does not attempt to parse/transform upstream args (composition over duplication).

Native mode:

- Default-deny for mutating operations: `pull` requires `--unsafe`.
- JSON encoding only (no string concatenation).
- Proxy handling is stricter: `no_proxy_auto=true` bypasses proxies for the configured host without mutating `NO_PROXY`.

Prompt handling:

- Prompts are treated as data (argv element or JSON field), not code.
- UI calls use `--` before the prompt when invoking the upstream CLI to avoid accidental flag parsing.

## UX Integration Strategy

- Same top-level commands in both modes (`list`, `ps`, `run`, ...).
- Same configuration keys (TOML, `.env`, environment variables, flags).
- Clear errors when a command is not supported in native mode.
- `doctor` reports:
  - effective host/lang/mode/unsafe
  - whether the upstream `ollama` CLI is available
  - whether the remote server responds to `/api/version`

## Rationale (Opinionated)

- Prefer wrapper mode when possible: upstream CLI stays the source of truth and minimizes drift.
- Keep native mode intentionally small: it is a fallback, not a reimplementation.
- Gate mutating operations in native mode: reduces surprise costs (downloads) and risk in a public, security-conscious tool.
- Fail fast on explicit executable configuration: avoids silent mode switching when the user clearly expected wrapper mode.
