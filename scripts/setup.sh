#!/bin/bash
# Ricochet Setup - one-time project setup (contexts, MCP, examples)

set -e

echo "🛠️  Ricochet Setup"

# Create default context from current folder
./scripts/detect-project-type.sh . || true
./scripts/create-context-from-folder.sh . || true

# Start MCP
./scripts/start-mcp.sh || true
./scripts/status-mcp.sh || true

echo "✅ Setup complete"
