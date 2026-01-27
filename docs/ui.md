# UI

Launch:

```bash
ollama-remote ui
```

If your Ollama server is not on the local machine, pass an explicit host (global flags must come first):

```bash
ollama-remote --host http://10.65.117.212:11434 ui
```

The UI:

- is optional (CLI remains primary)
- binds to `127.0.0.1` only
- uses the same hybrid runner as the CLI (wrapper mode when available, native mode as fallback)
- does not persist prompt history by default

Tip: the UI uses your effective config. If you want it to use the local Ollama CLI, set `mode=wrapper` and (if needed) `ollama_exe`.

## Rebuilding UI assets

The Go binary embeds `internal/ui/static/`. The generated JS bundle is `internal/ui/static/app.js`.

To rebuild after editing TypeScript:

```bash
go generate ./...
```

This requires Node.js (for `npx`) and will run esbuild against `internal/ui/frontend/app.ts`.
