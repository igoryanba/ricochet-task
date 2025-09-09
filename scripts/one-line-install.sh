#!/usr/bin/env bash
set -euo pipefail
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
cd "$TMP"
REPO="https://github.com/YOUR_ORG/ricochet-task.git"
echo "Fetching $REPO..."
if command -v git >/dev/null 2>&1; then
  git clone "$REPO"
  cd ricochet-task
else
  echo "❌ git is required"
  exit 1
fi
./scripts/install-cli.sh "$HOME/.local/bin"
echo "export PATH=\"$HOME/.local/bin:$PATH\"" >> "$HOME/.zshrc"
echo "✅ Installed. Restart your terminal or run: export PATH=\"$HOME/.local/bin:$PATH\""
