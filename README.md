# Mnemonic

Organizational knowledge management system with semantic search and progressive depth layers.

## Features

- **Semantic search** across 5 business domains (commercial, operations, financial, engineering, knowledge)
- **Progressive depth layers** — agents consume only the tokens they need
- **ChromaDB** remote vector storage with automatic embeddings
- **Dolibarr ERP sync** — direct REST API, delta incremental
- **Claude Code plugin** — hooks for proactive search and save
- **Single Go binary** — no Python, no CGO, no external dependencies

## Quick Start

```bash
# Install
go install github.com/marioser/mnemonic/cmd/mnemonic@latest

# Download embedding model (~30MB, one time)
mnemonic model download

# Configure
mnemonic init

# Sync from Dolibarr ERP
mnemonic sync-erp

# Start MCP server (used by Claude Code)
mnemonic mcp

# Start HTTP server (used by hooks)
mnemonic serve
```

## Architecture

```
mnemonic (Go binary)
├── MCP Server (stdio) — 25 tools for Claude Code
├── HTTP Server (:7438) — hooks, health, admin
├── Embedding Engine — all-MiniLM-L6-v2 (384 dims, pure Go)
├── ChromaDB Client — remote vector storage
└── Dolibarr Client — direct REST API sync
```

## The Onion Principle

Every piece of data has 4 layers of depth. Agents peel only what they need:

| Layer | Tokens | What | How |
|-------|--------|------|-----|
| 0 | ~50 | IDs, titles, metadata | `search_quick`, `browse` |
| 1 | ~200 | Semantic search + summary | `search`, `search_*` |
| 2 | ~500-2000 | Full document | `get_entity` |
| 3 | N × ~50 | Related entities | `find_related` |

## Domains

| Domain | Collection | Entity Types |
|--------|-----------|-------------|
| Commercial | mn-commercial | opportunity, proposal, client, competitor, client_comm, followup |
| Operations | mn-operations | project, task, delivery, timeline, quality, logistics |
| Financial | mn-financial | budget, apu, procurement, invoice, margin, expense |
| Engineering | mn-engineering | architecture, equipment, standard, protocol, config, concept |
| Knowledge | mn-knowledge | lesson, decision, conversation, agent_output, pattern |
| References | mn-references | reference, relationship, sync_state |

## License

MIT
