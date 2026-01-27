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

If you want native mode without changing config:

```bash
ollama-remote --mode native list
```

Or unset the configured path:

```bash
ollama-remote config set ollama_exe ""
```

## Connection refused on LAN IP (e.g. 10.65.117.212)

If Ollama works on the server machine via `http://127.0.0.1:11434` but fails via its LAN IP (for example `http://10.65.117.212:11434`), the Ollama server is likely bound to localhost only.

Quick checks:

```bash
curl http://127.0.0.1:11434/api/version
curl http://10.65.117.212:11434/api/version
```

To allow other machines on the LAN to connect:

- Configure Ollama to listen on `0.0.0.0:11434` (or the LAN interface)
- Allow inbound TCP `11434` in the firewall for your network

Then point `ollama-remote` at the LAN host:

```bash
ollama-remote --mode native --host http://10.65.117.212:11434 list
```

Security note: only do this if you trust the network (or add network controls), since it exposes the Ollama API to other machines.

## Proxy issues

If your environment uses a proxy and your host is a private address, you may need `NO_PROXY`.

- Recommended: set `NO_PROXY` yourself in your shell/profile
- Optional: set `no_proxy_auto = true` in config (applies only to the spawned process)
