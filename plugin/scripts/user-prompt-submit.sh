#!/usr/bin/env bash
# mn-user-prompt.sh — Mnemonic UserPromptSubmit hook
# First message: inject Knowledge Protocol
# Subsequent: auto-search KB when keywords detected, inject as additionalContext
set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"
SESSION_KEY="${CLAUDE_SESSION_ID:-default}"
STATE_FILE="/tmp/mnemonic-claude-${SESSION_KEY}-loaded"

# Read hook input from stdin
INPUT=$(cat 2>/dev/null || echo "{}")

# --- First message: inject Knowledge Protocol ---
if [ ! -f "$STATE_FILE" ]; then
  touch "$STATE_FILE"

  # Get KB status
  CONTEXT=""
  if curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
    RESP=$(curl -sf "${MNEMONIC_URL}/context" 2>/dev/null || echo "")
    if [ -n "$RESP" ]; then
      CONTEXT=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('summary',''))" 2>/dev/null || echo "")
    fi
  fi

  PROTOCOL="## Mnemonic Knowledge Base — ACTIVE

${CONTEXT}

### Tools by Layer (use lightest layer first):
**Layer 0** (~50 tok): search_quick, browse, count, list_types
**Layer 1** (~200 tok): search, search_commercial/operations/financial/engineering/knowledge
**Layer 2** (~500-2000 tok): get_entity, get_entities — ON DEMAND ONLY
**Layer 3** (graph): find_related, link_entities, get_timeline

### Proactive SEARCH (do without being asked):
- Before quoting → search_commercial + search_engineering
- Before estimating → search_financial(\"APU similar\")
- Client mentioned → search_quick(client=X)
- Similar project → search_operations(\"project similar to...\")

### Proactive SAVE (do without being asked):
- Technical decision → save_entity(domain=engineering, type=architecture)
- Lesson learned → save_entity(domain=knowledge, type=lesson)
- Client communication → save_entity(domain=commercial, type=client_comm)
- Equipment selection → save_entity(domain=engineering, type=equipment)

ALWAYS start with Layer 0 or 1. NEVER request Layer 2 for multiple entities."

  python3 -c "
import json
result = {
  'hookSpecificOutput': {
    'hookEventName': 'UserPromptSubmit',
    'additionalContext': '''$(echo "$PROTOCOL" | sed "s/'/\\\\'/g")'''
  }
}
print(json.dumps(result))
"
  exit 0
fi

# --- Subsequent messages: auto-search KB ---
if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
  echo '{}'
  exit 0
fi

# Extract user prompt from hook input
PROMPT=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('user_prompt',''))" 2>/dev/null || echo "")

if [ -z "$PROMPT" ]; then
  echo '{}'
  exit 0
fi

LOWER=$(echo "$PROMPT" | tr '[:upper:]' '[:lower:]')
RESULTS=""

# Detect quoting/proposal intent
if echo "$LOWER" | grep -qiE "cotizar|cotización|propuesta|presupuesto|quote|proposal"; then
  ENCODED=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1][:200]))" "$PROMPT" 2>/dev/null || echo "")
  if [ -n "$ENCODED" ]; then
    RESP=$(curl -sf "${MNEMONIC_URL}/hook/search?q=${ENCODED}&domain=commercial" 2>/dev/null || echo "")
    ITEMS=$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for r in data.get('results', []):
    print(f\"- [{r['type']}] {r['title']} (sim: {r.get('similarity',0):.0%})\")
" 2>/dev/null || echo "")
    if [ -n "$ITEMS" ]; then
      RESULTS="${RESULTS}KB commercial:\n${ITEMS}\n"
    fi
  fi
fi

# Detect technical keywords
if echo "$LOWER" | grep -qiE "plc|scada|hmi|rtu|dcs|variador|vfd|ups|tablero|instrumentación|arquitectura|diseño.*control"; then
  ENCODED=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1][:200]))" "$PROMPT" 2>/dev/null || echo "")
  if [ -n "$ENCODED" ]; then
    RESP=$(curl -sf "${MNEMONIC_URL}/hook/search?q=${ENCODED}&domain=engineering" 2>/dev/null || echo "")
    ITEMS=$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for r in data.get('results', []):
    print(f\"- [{r['type']}] {r['title']} (sim: {r.get('similarity',0):.0%})\")
" 2>/dev/null || echo "")
    if [ -n "$ITEMS" ]; then
      RESULTS="${RESULTS}KB engineering:\n${ITEMS}\n"
    fi
  fi
fi

# Detect project keywords
if echo "$LOWER" | grep -qiE "proyecto|cronograma|entrega|comisionamiento|puesta en marcha"; then
  ENCODED=$(python3 -c "import urllib.parse,sys; print(urllib.parse.quote(sys.argv[1][:200]))" "$PROMPT" 2>/dev/null || echo "")
  if [ -n "$ENCODED" ]; then
    RESP=$(curl -sf "${MNEMONIC_URL}/hook/search?q=${ENCODED}&domain=operations" 2>/dev/null || echo "")
    ITEMS=$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for r in data.get('results', []):
    print(f\"- [{r['type']}] {r['title']} (sim: {r.get('similarity',0):.0%})\")
" 2>/dev/null || echo "")
    if [ -n "$ITEMS" ]; then
      RESULTS="${RESULTS}KB operations:\n${ITEMS}\n"
    fi
  fi
fi

# Detect client names
CLIENT_MATCH=$(echo "$PROMPT" | grep -oiE "ecopetrol|drummond|cerrejon|reficar|promigas|transelca|celsia|argos|tecnoglass" | head -1 || echo "")
if [ -n "$CLIENT_MATCH" ]; then
  RESP=$(curl -sf "${MNEMONIC_URL}/hook/search?q=${CLIENT_MATCH}&domain=commercial" 2>/dev/null || echo "")
  ITEMS=$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for r in data.get('results', []):
    print(f\"- [{r['type']}] {r['title']}\")
" 2>/dev/null || echo "")
  if [ -n "$ITEMS" ]; then
    RESULTS="${RESULTS}KB ${CLIENT_MATCH}:\n${ITEMS}\n"
  fi
fi

# Inject results as additionalContext
if [ -n "$RESULTS" ]; then
  CONTEXT="Mnemonic auto-search results:\n${RESULTS}Use get_entity(id) for full details."
  python3 -c "
import json
result = {
  'hookSpecificOutput': {
    'hookEventName': 'UserPromptSubmit',
    'additionalContext': '''$(echo -e "$CONTEXT" | sed "s/'/\\\\'/g")'''
  }
}
print(json.dumps(result))
"
else
  echo '{}'
fi
