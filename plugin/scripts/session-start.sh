#!/usr/bin/env bash
# mn-session-start.sh — Ensure mnemonic serve is running at session start
set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"

# Start mnemonic serve if not running
if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
  mnemonic serve > /dev/null 2>&1 &
  sleep 2
fi

echo '{}'
