# Troubleshooting

## "Ollama CLI not found"

- Install Ollama and ensure `ollama` is available on PATH
- Or set `OLLAMA_EXE` / `ollama_exe` to the full path

## Proxy issues

If your environment uses a proxy and your host is a private address, you may need `NO_PROXY`.

- Recommended: set `NO_PROXY` yourself in your shell/profile
- Optional: set `no_proxy_auto = true` in config (applies only to the spawned process)
