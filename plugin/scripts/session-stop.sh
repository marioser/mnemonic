#!/usr/bin/env bash
# session-stop.sh — Save session summary and cleanup
# Captures what was worked on during the session and saves to KB.

set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"
SESSION_KEY="${CLAUDE_SESSION_ID:-default}"

# Clean up state files
rm -f "/tmp/mnemonic-claude-${SESSION_KEY}-loaded"
rm -f "/tmp/mnemonic-claude-${SESSION_KEY}-lastsearch"

# Check if mnemonic server is running
if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
  exit 0
fi

# Read session context from stdin if available (Claude Code passes conversation summary)
SUMMARY=""
if [ ! -t 0 ]; then
  SUMMARY=$(cat 2>/dev/null || true)
fi

# If we got a summary, save it
if [ -n "$SUMMARY" ]; then
  # Extract topics from first line or heading
  TOPICS=$(echo "$SUMMARY" | head -5 | tr '\n' ' ' | head -c 200)

  curl -sf "${MNEMONIC_URL}/hook/save-session-summary" \
    -X POST \
    -H "Content-Type: application/json" \
    -d "$(python3 -c "
import json
data = {
    'session_id': '${SESSION_KEY}',
    'summary': '''$(echo "$SUMMARY" | head -c 2000 | sed "s/'/\\\\'/g")''',
    'topics': '''$(echo "$TOPICS" | sed "s/'/\\\\'/g")'''
}
print(json.dumps(data))
" 2>/dev/null || echo '{"session_id":"","summary":"","topics":""}' )" \
    > /dev/null 2>&1 || true
fi

# Notify session end
curl -sf "${MNEMONIC_URL}/session/end" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "{\"session_id\": \"${SESSION_KEY}\"}" \
  > /dev/null 2>&1 || true
