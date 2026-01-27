# Security

## Reporting

Please open a private report (GitHub Security Advisories) if possible.

## Notes

- Do not include private hostnames, IPs, or tokens in issues.
- The optional UI binds to localhost only.

## Hybrid execution

- Wrapper mode spawns the official `ollama` CLI using argv arrays (no shell).
- Native mode calls the Ollama REST API directly; mutating operations are more restrictive by default.
- Configuration is validated; explicit `OLLAMA_EXE` misconfiguration fails fast instead of silently switching modes.
