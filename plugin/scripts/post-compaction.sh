#!/usr/bin/env bash
# post-compaction.sh — Reset first-message flag so protocol gets re-injected
set -euo pipefail
SESSION_KEY="${CLAUDE_SESSION_ID:-default}"
rm -f "/tmp/mnemonic-claude-${SESSION_KEY}-loaded"
echo '{}'
