#!/bin/bash
# Ricochet Installer - setup dependencies, config, and hooks

set -e

echo "🚀 Ricochet Installer"

# 1) Ensure scripts executable
chmod +x scripts/*.sh || true
chmod +x scripts/tests/*.sh || true

# 2) Verify dependencies
need() { command -v "$1" >/dev/null 2>&1 || { echo "❌ Missing dependency: $1"; MISSING=1; }; }
MISSING=0
for bin in jq curl git; do need "$bin"; done
[ "$MISSING" = "1" ] && { echo "Install missing deps and re-run."; exit 1; }

# 3) MCP config hint
if [ ! -f "$HOME/.cursor/mcp.json" ]; then
  echo "⚠️  No ~/.cursor/mcp.json found. Create it to enable MCP in Cursor."
fi

# 4) Git hooks
./scripts/ai-git-hooks.sh install . || true

# 5) Validate ricochet config
if [ -f "ricochet.yaml" ]; then
  ./ricochet-task config validate || true
else
  echo "⚠️  ricochet.yaml not found. Create it from example."
fi

echo "✅ Install complete"
