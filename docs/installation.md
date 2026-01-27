# Installation

This project ships a single CLI binary: `ollama-remote`.

## Windows (recommended)

1) Download the latest release zip.
2) Extract it to a folder (example: `C:\\Tools\\ollama-remote`).
3) Add that folder to your User PATH:

```powershell
pwsh -NoProfile -ExecutionPolicy Bypass -File .\install.ps1
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
