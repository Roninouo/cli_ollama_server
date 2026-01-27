#!/usr/bin/env sh
set -eu

# Adds this folder to PATH in your shell profile.
#
# Usage:
#   ./install.sh
#   ./install.sh --dry-run

dry_run=0
if [ "${1-}" = "--dry-run" ] || [ "${1-}" = "-n" ]; then
  dry_run=1
fi

here_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

# Prefer a file that is commonly sourced.
profile=""
if [ -n "${ZSH_VERSION-}" ] && [ -f "$HOME/.zshrc" ]; then
  profile="$HOME/.zshrc"
elif [ -n "${BASH_VERSION-}" ] && [ -f "$HOME/.bashrc" ]; then
  profile="$HOME/.bashrc"
elif [ -f "$HOME/.profile" ]; then
  profile="$HOME/.profile"
elif [ -f "$HOME/.bash_profile" ]; then
  profile="$HOME/.bash_profile"
else
  profile="$HOME/.profile"
fi

line="export PATH=\"$here_dir:\$PATH\""

if [ $dry_run -eq 1 ]; then
  printf '%s\n' "Dry run (no changes made)."
  printf '%s\n' "Would add to PATH in: $profile"
  printf '%s\n' "$line"
  exit 0
fi

mkdir -p "$(dirname -- "$profile")"
touch "$profile"

if grep -Fq "$here_dir" "$profile" 2>/dev/null; then
  printf '%s\n' "Already on PATH via: $profile"
  exit 0
fi

{
  printf '\n# Added by ollama-remote installer\n'
  printf '%s\n' "$line"
} >>"$profile"

printf '%s\n' "Added to PATH via: $profile"
printf '%s\n' "Restart your shell (or run: . \"$profile\")"
