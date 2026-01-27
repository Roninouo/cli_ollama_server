# Installation

This project ships a single CLI binary: `ollama-remote`.

It does not ship the upstream Ollama CLI or server.

- If you want wrapper mode (`--mode wrapper`), install Ollama locally so `ollama` / `ollama.exe` exists.
- If you do not want a local Ollama install, use native mode (`--mode native`) for supported commands.

## Windows (recommended)

1) Download the latest release zip.
2) Extract it to a folder (example: `C:\\Tools\\ollama-remote`).
3) Add that folder to your User PATH:

```powershell
pwsh -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
```

Optional: verify the local Ollama CLI is available:

```powershell
where.exe ollama
ollama --version
```

If it is not on PATH, configure `ollama-remote` with an explicit path:

```powershell
ollama-remote config set ollama_exe "C:\Program Files\Ollama\ollama.exe"
ollama-remote config set mode wrapper
```

## macOS / Linux

1) Download the latest release zip for your OS.
2) Extract it to a folder (example: `~/tools/ollama-remote`).
3) Add that folder to your PATH:

```sh
chmod +x ./install.sh ./ollama-remote
./install.sh
```

To remove it from your PATH:

```sh
./uninstall.sh
```

## From source

Requires Go 1.22+.

```powershell
go build -o .\dist\ollama-remote.exe .\cmd\ollama-remote
```
