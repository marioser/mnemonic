#!/usr/bin/env bash
# subagent-stop.sh — Capture knowledge from sub-agent output
# Runs async after each sub-agent completes.

set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"

# Read sub-agent output from stdin
OUTPUT=""
if [ ! -t 0 ]; then
  OUTPUT=$(cat 2>/dev/null || true)
fi

if [ -z "$OUTPUT" ]; then
  exit 0
fi

# Check if mnemonic server is running
if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
  exit 0
fi

# Extract key learnings sections from the output
# Look for patterns like "## Key Learnings:", "### Decisiones:", "**Learned:**"
if echo "$OUTPUT" | grep -qiE "key learning|decisión|decisiones|learned|lección|arquitectura elegida|equipo seleccionado"; then
  # Post to mnemonic for passive capture (future endpoint)
  # For now, the agent should call save_entity directly via MCP tools
  :
fi
