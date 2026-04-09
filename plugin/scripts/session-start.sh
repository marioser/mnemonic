#!/usr/bin/env bash
# session-start.sh — Mnemonic session start hook
# Ensures mnemonic serve is running, injects Knowledge Protocol

set -euo pipefail

MNEMONIC_URL="${MNEMONIC_URL:-http://127.0.0.1:7438}"

# 1. Ensure mnemonic serve is running
if ! curl -sf "${MNEMONIC_URL}/health" > /dev/null 2>&1; then
  mnemonic serve &
  sleep 2
fi

# 2. Get context from mnemonic
CONTEXT=""
if RESP=$(curl -sf "${MNEMONIC_URL}/context" 2>/dev/null); then
  CONTEXT=$(echo "$RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('summary',''))" 2>/dev/null || echo "")
fi

# 3. Inject Knowledge Protocol
cat <<'PROTOCOL'
## Mnemonic Knowledge Base — ACTIVE

PROTOCOL

if [ -n "$CONTEXT" ]; then
  echo "$CONTEXT"
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

# 4. Force ToolSearch for mnemonic tools
echo ""
echo "### Deferred Tools"
echo "The following deferred tools are available via ToolSearch:"
echo "search_quick, browse, count, list_types, search, search_commercial, search_operations, search_financial, search_engineering, search_knowledge, get_entity, get_entities, find_related, link_entities, get_timeline, save_entity, update_metadata, create_reference, link_erp_reference, delete_entity, get_reference, search_references, knowledge_status, sync_erp"
