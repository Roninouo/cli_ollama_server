#!/usr/bin/env sh
set -eu

# Removes this folder from PATH lines in your shell profile.
#
# Usage:
#   ./uninstall.sh
#   ./uninstall.sh --dry-run

dry_run=0
if [ "${1-}" = "--dry-run" ] || [ "${1-}" = "-n" ]; then
  dry_run=1
fi

here_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)

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

if [ ! -f "$profile" ]; then
  printf '%s\n' "Profile not found: $profile"
  exit 0
fi

if [ $dry_run -eq 1 ]; then
  printf '%s\n' "Dry run (no changes made)."
  printf '%s\n' "Would remove lines containing: $here_dir"
  printf '%s\n' "From: $profile"
  exit 0
fi

tmp="$profile.ollama-remote.tmp"
awk -v needle="$here_dir" 'index($0, needle) == 0 { print }' "$profile" >"$tmp"
mv "$tmp" "$profile"

printf '%s\n' "Removed PATH entry (if present) from: $profile"
printf '%s\n' "Restart your shell to pick up the change"
