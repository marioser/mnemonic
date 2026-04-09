#!/usr/bin/env bash
# mn-subagent-stop.sh — Capture knowledge from sub-agent output
set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"

if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
  echo '{}'
  exit 0
fi

INPUT=$(cat 2>/dev/null || echo "{}")
OUTPUT=$(echo "$INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('tool_output','')[:2000])
except:
    pass
" 2>/dev/null || echo "")

if [ -z "$OUTPUT" ]; then
  echo '{}'
  exit 0
fi

AGENT_TYPE="unknown"
if echo "$OUTPUT" | grep -qiE "scope|alcance"; then AGENT_TYPE="scope-architect"
elif echo "$OUTPUT" | grep -qiE "cost|presupuesto|costo"; then AGENT_TYPE="cost-estimator"
elif echo "$OUTPUT" | grep -qiE "bom|materiales|bill of materials"; then AGENT_TYPE="bom-builder"
elif echo "$OUTPUT" | grep -qiE "sales|estrategia.*comercial"; then AGENT_TYPE="sales-strategist"
elif echo "$OUTPUT" | grep -qiE "roi|retorno"; then AGENT_TYPE="roi-calculator"
fi

if [ "$AGENT_TYPE" = "unknown" ]; then
  echo '{}'
  exit 0
fi

curl -sf "${MNEMONIC_URL}/hook/save-agent-output" -X POST -H "Content-Type: application/json" \
  -d "$(python3 -c "
import json
print(json.dumps({
    'content': '''$(echo "$OUTPUT" | head -c 1000 | sed "s/'/\\\\'/g")''',
    'agent_type': '$AGENT_TYPE',
    'session_id': '${CLAUDE_SESSION_ID:-default}'
}))
" 2>/dev/null)" > /dev/null 2>&1 || true

echo '{}'
