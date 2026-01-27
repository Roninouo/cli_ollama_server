# Troubleshooting

## "Ollama CLI not found"

- Install Ollama and ensure `ollama` is available on PATH
- Or set `OLLAMA_EXE` / `ollama_exe` to the full path

Windows tips:

```powershell
where.exe ollama
Get-Command ollama -ErrorAction SilentlyContinue
```

Common install path:

- `C:\Program Files\Ollama\ollama.exe`

If you do not want to install the CLI, use native mode for supported commands:

```bash
ollama-remote --mode native list
```

## "Invalid --ollama-exe / OLLAMA_EXE"

If you set `--ollama-exe` / `OLLAMA_EXE` but the executable cannot be resolved (and `mode != native`), `ollama-remote` fails fast.

- Fix the path, or unset it to allow `mode=auto` to fall back to native mode

## Proxy issues

If your environment uses a proxy and your host is a private address, you may need `NO_PROXY`.

- Recommended: set `NO_PROXY` yourself in your shell/profile
- Optional: set `no_proxy_auto = true` in config (applies only to the spawned process)
