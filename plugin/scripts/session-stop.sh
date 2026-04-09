#!/usr/bin/env bash
# mn-session-stop.sh — Cleanup on session end
set -euo pipefail

SESSION_KEY="${CLAUDE_SESSION_ID:-default}"
rm -f "/tmp/mnemonic-claude-${SESSION_KEY}-loaded"

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"
curl -sf "${MNEMONIC_URL}/session/end" -X POST -H "Content-Type: application/json" \
  -d "{\"session_id\": \"${SESSION_KEY}\"}" > /dev/null 2>&1 || true

echo '{}'
