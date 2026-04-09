#!/usr/bin/env bash
# session-stop.sh — Cleanup on session end

set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"
SESSION_KEY="${CLAUDE_SESSION_ID:-default}"

# Clean up state file
rm -f "/tmp/mnemonic-claude-${SESSION_KEY}-loaded"

# Notify mnemonic server
curl -sf "${MNEMONIC_URL}/session/end" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "{\"session_id\": \"${SESSION_KEY}\"}" \
  > /dev/null 2>&1 || true
