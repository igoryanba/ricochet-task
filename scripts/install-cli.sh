#!/usr/bin/env bash
set -euo pipefail

TARGET_NAME="ricochet-task"
PREFIX_DEFAULT="/usr/local/bin"
PREFIX="${1:-$PREFIX_DEFAULT}"
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BIN_SRC="$REPO_ROOT/$TARGET_NAME"
BIN_DST="$PREFIX/$TARGET_NAME"

echo "🚀 Installing $TARGET_NAME to $BIN_DST"

# 1) Build from source if no binary present
if [ ! -x "$BIN_SRC" ]; then
  if command -v go >/dev/null 2>&1; then
    echo "🔧 Building from source..."
    (cd "$REPO_ROOT" && go build -o "$TARGET_NAME" .)
  else
    echo "❌ No executable binary and Go not found. Install Go or place built binary at $BIN_SRC"
    exit 1
  fi
fi

# 2) Create target dir
mkdir -p "$PREFIX"

# 3) Install (copy or symlink)
if cp "$BIN_SRC" "$BIN_DST" 2>/dev/null; then
  echo "✅ Copied binary to $BIN_DST"
else
  echo "ℹ️  Copy failed, trying symlink (may need sudo)"
  ln -sf "$BIN_SRC" "$BIN_DST"
  echo "✅ Symlinked $BIN_DST -> $BIN_SRC"
fi

# 4) Verify
if command -v "$TARGET_NAME" >/dev/null 2>&1; then
  echo "✅ $TARGET_NAME available: $(command -v $TARGET_NAME)"
  "$TARGET_NAME" --help >/dev/null 2>&1 || true
else
  echo "⚠️  $TARGET_NAME not on PATH. Add $PREFIX to PATH or provide a different prefix"
fi

echo "Done."
