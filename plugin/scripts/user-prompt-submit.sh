#!/usr/bin/env bash
# user-prompt-submit.sh — Mnemonic prompt hook
# First message: force ToolSearch. Subsequent: contextual nudge.

set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"
SESSION_KEY="${CLAUDE_SESSION_ID:-default}"
STATE_FILE="/tmp/mnemonic-claude-${SESSION_KEY}-loaded"

# Read user prompt from stdin
PROMPT=""
if [ -t 0 ]; then
  : # no stdin
else
  PROMPT=$(cat 2>/dev/null || true)
fi

# First message detection
if [ ! -f "$STATE_FILE" ]; then
  touch "$STATE_FILE"
  # Force load mnemonic tools
  echo "select:mcp__mnemonic__search_quick,mcp__mnemonic__search,mcp__mnemonic__search_commercial,mcp__mnemonic__search_operations,mcp__mnemonic__search_financial,mcp__mnemonic__search_engineering,mcp__mnemonic__search_knowledge,mcp__mnemonic__get_entity,mcp__mnemonic__get_entities,mcp__mnemonic__find_related,mcp__mnemonic__save_entity,mcp__mnemonic__knowledge_status"
  exit 0
fi

# Contextual nudge based on prompt keywords
if [ -n "$PROMPT" ]; then
  LOWER=$(echo "$PROMPT" | tr '[:upper:]' '[:lower:]')

  if echo "$LOWER" | grep -qiE "cotizar|propuesta|presupuesto|quote|proposal"; then
    echo "Tip: Search the KB before quoting — use search_commercial and search_engineering for context."
  elif echo "$LOWER" | grep -qiE "plc|scada|equipo|arquitectura|diseño|hmi|rtu"; then
    echo "Tip: Check prior technical decisions — use search_engineering."
  elif echo "$LOWER" | grep -qiE "cliente|ecopetrol|isa |drummond|cerrejon"; then
    echo "Tip: Look up client history — use search_quick with client filter."
  fi
fi
