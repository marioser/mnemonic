#!/usr/bin/env bash
# session-start.sh — Mnemonic session start hook
# 1. Ensures mnemonic serve is running
# 2. Loads recent activity from KB
# 3. Injects Knowledge Protocol with live context

set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"

# --- 1. Ensure mnemonic serve is running ---
if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
  mnemonic serve &
  sleep 2
  if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
    echo "Mnemonic KB: server not available. Tools will work but hooks are limited."
    exit 0
  fi
fi

# --- 2. Get KB status ---
CONTEXT=""
if RESP=$(curl -sf "${MNEMONIC_URL}/context" 2>/dev/null); then
  CONTEXT=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('summary',''))" 2>/dev/null || echo "")
fi

# --- 3. Get recent activity per domain ---
RECENT_COMMERCIAL=""
RECENT_OPERATIONS=""
RECENT_ENGINEERING=""

for DOMAIN in commercial operations engineering; do
  RESP=$(curl -sf "${MNEMONIC_URL}/hook/recent?domain=${DOMAIN}" 2>/dev/null || echo "")
  if [ -n "$RESP" ]; then
    COUNT=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('count',0))" 2>/dev/null || echo "0")
    if [ "$COUNT" -gt 0 ] 2>/dev/null; then
      ITEMS=$(echo "$RESP" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for r in data.get('results', [])[:3]:
    print(f\"  - {r['type']}: {r['title']}\")
" 2>/dev/null || echo "")
      case "$DOMAIN" in
        commercial)  RECENT_COMMERCIAL="$ITEMS" ;;
        operations)  RECENT_OPERATIONS="$ITEMS" ;;
        engineering) RECENT_ENGINEERING="$ITEMS" ;;
      esac
    fi
  fi
done

# --- 4. Inject Knowledge Protocol ---
cat <<'PROTOCOL'
## Mnemonic Knowledge Base — ACTIVE

PROTOCOL

if [ -n "$CONTEXT" ]; then
  echo "$CONTEXT"
  echo ""
fi

# Inject recent activity if available
HAS_RECENT=false
if [ -n "$RECENT_COMMERCIAL" ] || [ -n "$RECENT_OPERATIONS" ] || [ -n "$RECENT_ENGINEERING" ]; then
  HAS_RECENT=true
  echo "### Recent KB Activity"
fi

if [ -n "$RECENT_COMMERCIAL" ]; then
  echo "**Commercial:**"
  echo "$RECENT_COMMERCIAL"
fi
if [ -n "$RECENT_OPERATIONS" ]; then
  echo "**Operations:**"
  echo "$RECENT_OPERATIONS"
fi
if [ -n "$RECENT_ENGINEERING" ]; then
  echo "**Engineering:**"
  echo "$RECENT_ENGINEERING"
fi
if [ "$HAS_RECENT" = true ]; then
  echo ""
fi

cat <<'PROTOCOL'
### Tools by Layer (use lightest layer first):

**Layer 0 — Inventory (no embeddings, ~50 tokens/result):**
- `search_quick` — Fast metadata search. Use FIRST.
- `browse` — Paginated listing of a domain
- `count` — Entity counts per domain
- `list_types` — Available entity types

**Layer 1 — Semantic Search (~200 tokens/result):**
- `search` — Cross-domain semantic search
- `search_commercial` / `search_operations` / `search_financial` / `search_engineering` / `search_knowledge`

**Layer 2 — Full Detail (~500-2000 tokens, on demand only):**
- `get_entity` — Complete document for one entity
- `get_entities` — Batch get (max 10)

**Layer 3 — Relationships (graph traversal):**
- `find_related` — Connected entities
- `link_entities` — Create relationship
- `get_timeline` — Chronological history

### When to SEARCH (do it proactively):
- Before quoting → search_commercial + search_engineering
- Before estimating costs → search_financial("APU similar")
- Before choosing equipment → search_engineering("equipment selection")
- Client mentioned → search_quick(client=X) for context
- Similar project → search_operations("project similar to...")

### When to SAVE (do it proactively):
- Technical decision → save_entity(domain=engineering, type=architecture)
- Lesson learned → save_entity(domain=knowledge, type=lesson)
- Client communication → save_entity(domain=commercial, type=client_comm)
- Equipment selection → save_entity(domain=engineering, type=equipment)
- Proposal generated → create_reference + save_entity

### ALWAYS start with Layer 0 or 1. NEVER request Layer 2 for multiple entities at once.
PROTOCOL
