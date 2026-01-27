# UI

Launch:

```bash
ollama-remote ui
```

The UI:

- is optional (CLI remains primary)
- binds to `127.0.0.1` only
- uses the same hybrid runner as the CLI (wrapper mode when available, native mode as fallback)
- does not persist prompt history by default
