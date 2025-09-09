#!/usr/bin/env bash
# Safe wrapper around ./ricochet-task with friendly provider/health diagnostics
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CLI="$REPO_ROOT/ricochet-task"

run_with_help() {
  cmd=("$CLI" "$@")
  # capture both streams
  out="$(${cmd[@]} 2>&1 || true)"
  ec=$?
  echo "$out"
  if [ $ec -ne 0 ]; then
    if echo "$out" | grep -qi "provider not found\|Failed to get provider"; then
      echo "\n⚠️  Provider issue detected. Quick checks:" >&2
      echo "   1) Verify provider name and case in ricochet.yaml (e.g. youtrack)" >&2
      echo "   2) Ensure project/board IDs exist (e.g. --project \"0-1\" --board \"0-2\")" >&2
      echo "   3) Verify API token is set and valid" >&2
      echo "   4) Run health checks:" >&2
      echo "      - $CLI status" >&2
      echo "      - $CLI provider test youtrack" >&2
    fi
    exit $ec
  fi
}

case "${1:-}" in
  tasks)
    shift
    case "${1:-}" in
      list|update|create|get)
        # Pre-flight: quick status (non-fatal)
        "$CLI" status >/dev/null 2>&1 || true
        run_with_help tasks "$@"
        ;;
      *)
        exec "$CLI" tasks "$@"
        ;;
    esac
    ;;
  *)
    exec "$CLI" "$@"
    ;;
esac
