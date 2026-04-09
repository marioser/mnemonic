#!/usr/bin/env bash
# user-prompt-submit.sh — Mnemonic prompt hook
# 1. First message: force ToolSearch to load mnemonic tools
# 2. Subsequent: detect clients/projects/tech keywords → auto-search KB → inject context

set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"
SESSION_KEY="${CLAUDE_SESSION_ID:-default}"
STATE_FILE="/tmp/mnemonic-claude-${SESSION_KEY}-loaded"
LAST_SEARCH_FILE="/tmp/mnemonic-claude-${SESSION_KEY}-lastsearch"

# Read user prompt from stdin
PROMPT=""
if [ ! -t 0 ]; then
  PROMPT=$(cat 2>/dev/null || true)
fi

# --- First message: load tools ---
if [ ! -f "$STATE_FILE" ]; then
  touch "$STATE_FILE"
  exit 0
fi

# --- No prompt or server not running → skip ---
if [ -z "$PROMPT" ]; then
  exit 0
fi

if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
  exit 0
fi

# --- Detect keywords and auto-search ---
LOWER=$(echo "$PROMPT" | tr '[:upper:]' '[:lower:]')
RESULTS=""

# Detect client names (common patterns)
CLIENT_MATCH=""
if echo "$LOWER" | grep -qiE "ecopetrol|drummond|cerrejon|reficar|promigas|transelca|celsia|isa\b|argos|tecnoglass|bimbo|nutresa"; then
  # Extract the matched client name
  CLIENT_MATCH=$(echo "$PROMPT" | grep -oiE "ecopetrol|drummond|cerrejon|reficar|promigas|transelca|celsia|isa|argos|tecnoglass|bimbo|nutresa" | head -1)
fi

# Detect quoting/proposal intent
if echo "$LOWER" | grep -qiE "cotizar|cotización|propuesta|presupuesto|quote|proposal|precio"; then
  RESP=$(curl -sf "${MNEMONIC_URL}/hook/search?q=$(echo "$PROMPT" | head -c 200 | python3 -c 'import sys,urllib.parse; print(urllib.parse.quote(sys.stdin.read().strip()))')&domain=commercial" 2>/dev/null || echo "")
  COUNT=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('count',0))" 2>/dev/null || echo "0")
  if [ "$COUNT" -gt 0 ] 2>/dev/null; then
    RESULTS="${RESULTS}$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print('**KB — propuestas/clientes relacionados:**')
for r in data.get('results', []):
    sim = r.get('similarity', 0)
    print(f\"  - [{r['type']}] {r['title']} (sim: {sim:.0%})\")
" 2>/dev/null || echo "")
"
  fi
fi

# Detect technical/engineering keywords
if echo "$LOWER" | grep -qiE "plc|scada|hmi|rtu|dcs|variador|vfd|ups|tablero|instrumentación|sensores|arquitectura.*técnica|diseño.*control"; then
  QUERY=$(echo "$PROMPT" | head -c 200 | python3 -c 'import sys,urllib.parse; print(urllib.parse.quote(sys.stdin.read().strip()))')
  RESP=$(curl -sf "${MNEMONIC_URL}/hook/search?q=${QUERY}&domain=engineering" 2>/dev/null || echo "")
  COUNT=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('count',0))" 2>/dev/null || echo "0")
  if [ "$COUNT" -gt 0 ] 2>/dev/null; then
    RESULTS="${RESULTS}$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print('**KB — conocimiento técnico relacionado:**')
for r in data.get('results', []):
    sim = r.get('similarity', 0)
    print(f\"  - [{r['type']}] {r['title']} (sim: {sim:.0%})\")
" 2>/dev/null || echo "")
"
  fi
fi

# Detect project/operations keywords
if echo "$LOWER" | grep -qiE "proyecto|cronograma|entrega|tarea|milestone|wbs|gantt|comisionamiento|puesta en marcha"; then
  QUERY=$(echo "$PROMPT" | head -c 200 | python3 -c 'import sys,urllib.parse; print(urllib.parse.quote(sys.stdin.read().strip()))')
  RESP=$(curl -sf "${MNEMONIC_URL}/hook/search?q=${QUERY}&domain=operations" 2>/dev/null || echo "")
  COUNT=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('count',0))" 2>/dev/null || echo "0")
  if [ "$COUNT" -gt 0 ] 2>/dev/null; then
    RESULTS="${RESULTS}$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print('**KB — proyectos relacionados:**')
for r in data.get('results', []):
    sim = r.get('similarity', 0)
    print(f\"  - [{r['type']}] {r['title']} (sim: {sim:.0%})\")
" 2>/dev/null || echo "")
"
  fi
fi

# Client-specific search
if [ -n "$CLIENT_MATCH" ]; then
  RESP=$(curl -sf "${MNEMONIC_URL}/hook/search?q=${CLIENT_MATCH}&domain=commercial" 2>/dev/null || echo "")
  COUNT=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('count',0))" 2>/dev/null || echo "0")
  if [ "$COUNT" -gt 0 ] 2>/dev/null; then
    RESULTS="${RESULTS}$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(f'**KB — historial de {data[\"query\"]}:**')
for r in data.get('results', []):
    print(f\"  - [{r['type']}] {r['title']}\")
" 2>/dev/null || echo "")
"
  fi
fi

# --- Inject results if any ---
if [ -n "$RESULTS" ]; then
  echo "### Mnemonic KB Context (auto-searched)"
  echo "$RESULTS"
  echo ""
  echo "_Use get_entity(id) for full details on any result._"
fi
