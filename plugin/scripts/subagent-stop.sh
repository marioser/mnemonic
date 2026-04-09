#!/usr/bin/env bash
# subagent-stop.sh — Capture knowledge from sub-agent output
# Parses agent output for decisions, lessons, architectures, and saves to KB automatically.
# Runs async after each sub-agent completes.

set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"
SESSION_KEY="${CLAUDE_SESSION_ID:-default}"

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

# Detect agent type from output patterns
AGENT_TYPE="unknown"
if echo "$OUTPUT" | grep -qiE "scope.*architect|alcance|scope"; then
  AGENT_TYPE="scope-architect"
elif echo "$OUTPUT" | grep -qiE "cost.*estimat|presupuesto|costo total"; then
  AGENT_TYPE="cost-estimator"
elif echo "$OUTPUT" | grep -qiE "bom.*builder|materiales|lista de materiales|bill of materials"; then
  AGENT_TYPE="bom-builder"
elif echo "$OUTPUT" | grep -qiE "sales.*strateg|estrategia.*comercial|sales strategy"; then
  AGENT_TYPE="sales-strategist"
elif echo "$OUTPUT" | grep -qiE "roi.*calcul|retorno.*inversión|roi"; then
  AGENT_TYPE="roi-calculator"
elif echo "$OUTPUT" | grep -qiE "time.*estimat|cronograma|pert|duración"; then
  AGENT_TYPE="time-estimator"
elif echo "$OUTPUT" | grep -qiE "resource.*estimat|recursos.*humanos|equipo de trabajo"; then
  AGENT_TYPE="resource-estimator"
elif echo "$OUTPUT" | grep -qiE "proposal.*scor|scoring|evaluación|calificación"; then
  AGENT_TYPE="proposal-scorer"
fi

# Extract meaningful content (first 2000 chars, skip boilerplate)
CONTENT=$(echo "$OUTPUT" | head -c 2000)

# Only save if we detected a known agent type (skip noise)
if [ "$AGENT_TYPE" = "unknown" ]; then
  exit 0
fi

# Save to KB via HTTP API
curl -sf "${MNEMONIC_URL}/hook/save-agent-output" \
  -X POST \
  -H "Content-Type: application/json" \
  -d "$(python3 -c "
import json, sys
data = {
    'content': '''$(echo "$CONTENT" | sed "s/'/\\\\'/g")''',
    'agent_type': '${AGENT_TYPE}',
    'session_id': '${SESSION_KEY}'
}
print(json.dumps(data))
" 2>/dev/null || echo '{"content":"","agent_type":"unknown","session_id":""}' )" \
  > /dev/null 2>&1 || true
