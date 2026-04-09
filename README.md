# Mnemonic

Organizational knowledge management system with semantic search and progressive depth layers.

Built in Go. Designed for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

```
mnemonic mcp              # MCP server for Claude Code
mnemonic serve             # HTTP server for hooks
mnemonic sync-erp          # Sync from Dolibarr ERP
mnemonic sync-erp --client="Ecopetrol"  # Deep sync one client
mnemonic status            # KB status
```

---

## What is Mnemonic?

Mnemonic is a knowledge base that stores, searches, and relates business data across your organization. It connects to Claude Code as a plugin, giving AI agents access to your clients, projects, proposals, financials, engineering decisions, and lessons learned.

**Key features:**
- **Semantic search** — find things by meaning, not just keywords
- **5 business domains** — commercial, operations, financial, engineering, knowledge
- **Progressive depth** — agents consume only the tokens they need (The Onion Principle)
- **Dolibarr ERP sync** — direct REST API, incremental delta
- **Claude Code plugin** — hooks for proactive search and save
- **Single Go binary** — no Python, no external dependencies

---

## The Onion Principle

Every piece of data has 4 layers of depth. Agents peel only what they need:

```
┌───────────────────────────────────────────────────────┐
│ Layer 0 — Inventory (~50 tokens/result)               │
│ "What exists?" — IDs, titles, metadata                │
│ No embeddings. Instant.                               │
├───────────────────────────────────────────────────────┤
│ Layer 1 — Discovery (~200 tokens/result)              │
│ "What's relevant?" — Semantic search + summary        │
│ Uses embeddings. Finds by meaning.                    │
├───────────────────────────────────────────────────────┤
│ Layer 2 — Detail (~500-2000 tokens)                   │
│ "Tell me everything" — Full document + metadata       │
│ Only on demand, for specific entities.                │
├───────────────────────────────────────────────────────┤
│ Layer 3 — Context (~N × 50 tokens)                    │
│ "What's connected?" — Graph traversal                 │
│ Related entities, timeline, relationships.            │
└───────────────────────────────────────────────────────┘
```

An agent searching for "similar SCADA projects" gets Layer 0+1 results. If it needs more detail on a specific project, it requests Layer 2 for that one entity. **Never 2000 tokens for 10 projects when 50 tokens each was enough.**

---

## Architecture

```
┌──────────────────────────────────────────────────────┐
│               mnemonic (Go binary)                    │
│                                                       │
│  ┌─────────────┐          ┌──────────────────┐       │
│  │  MCP Server │          │   HTTP Server     │       │
│  │  (stdio)    │          │   (:7438)         │       │
│  │  25 tools   │          │   /health         │       │
│  │             │          │   /status          │       │
│  │  Claude     │          │   /context         │       │
│  │  Code ←→    │          │   Hooks ←→         │       │
│  └──────┬──────┘          └────────┬──────────┘       │
│         └──────────┬───────────────┘                  │
│                    │                                  │
│         ┌──────────▼──────────┐                       │
│         │  ChromaDB Client    │                       │
│         │  6 collections      │                       │
│         │  Cosine similarity  │                       │
│         └──────────┬──────────┘                       │
│                    │                                  │
│         ┌──────────▼──────────┐                       │
│         │  Dolibarr Client    │                       │
│         │  REST API direct    │                       │
│         │  Delta sync         │                       │
│         └─────────────────────┘                       │
└──────────────────────────────────────────────────────┘
                     │
                     ▼
          ChromaDB Server (remote)
          Dolibarr ERP (remote)
```

---

## Installation

### Prerequisites

- **Go 1.24+** — [Download](https://go.dev/dl/)
- **ChromaDB server** — running and accessible (local or remote)
- **Claude Code** — for the plugin integration

### Option A: From source (recommended for development)

```bash
# Clone
git clone https://github.com/marioser/mnemonic.git
cd mnemonic

# Build
go build -o mnemonic ./cmd/mnemonic/

# Install to PATH
cp mnemonic /usr/local/bin/
# or
cp mnemonic /opt/homebrew/bin/  # macOS with Homebrew
```

### Option B: Go install

```bash
go install github.com/marioser/mnemonic/cmd/mnemonic@latest
```

### Option C: Download binary

Download the latest release for your platform from [GitHub Releases](https://github.com/marioser/mnemonic/releases).

```bash
# macOS Apple Silicon
curl -L https://github.com/marioser/mnemonic/releases/latest/download/mnemonic_darwin_arm64.tar.gz | tar xz
cp mnemonic /usr/local/bin/

# macOS Intel
curl -L https://github.com/marioser/mnemonic/releases/latest/download/mnemonic_darwin_amd64.tar.gz | tar xz

# Linux x64
curl -L https://github.com/marioser/mnemonic/releases/latest/download/mnemonic_linux_amd64.tar.gz | tar xz
```

### Verify installation

```bash
mnemonic version
# mnemonic v0.1.0 (darwin/arm64)
```

---

## Configuration

Mnemonic looks for configuration in this priority order:

1. `MN_CONFIG` environment variable (highest priority)
2. `./config/mnemonic.yaml` (project-local)
3. `~/.mnemonic/config.yaml` (global)
4. Built-in defaults

### Create your config file

```bash
mkdir -p ~/.mnemonic
cat > ~/.mnemonic/config.yaml << 'EOF'
# ChromaDB server connection
chromadb:
  host: "localhost"        # Your ChromaDB server IP or hostname
  port: 8000
  token: ""                # Bearer token (leave empty if no auth)
  ssl: false
  collection_prefix: "mn"  # Collections: mn-commercial, mn-operations, etc.

# HTTP server for hooks
server:
  port: 7438
  host: "127.0.0.1"

# Search defaults
search:
  default_results: 5
  min_similarity: 0.7

# Dolibarr ERP (optional — only needed for sync-erp)
dolibarr:
  url: ""                  # e.g. "https://your-dolibarr.com"
  api_key: ""              # Dolibarr API key (DOLAPIKEY)
  sync:
    delta_days: 365
    batch_size: 100
    entities:
      customers: true
      projects: true
      proposals: true
      products: true

# Logging
log:
  level: "info"            # debug, info, warn, error
  format: "text"           # text, json
EOF
```

### Environment variables

Any config value can be overridden with environment variables:

| Variable | Config path | Example |
|----------|------------|---------|
| `MN_CONFIG` | — | Path to config file |
| `MNEMONIC_CHROMADB_HOST` | `chromadb.host` | `192.168.1.100` |
| `MNEMONIC_CHROMADB_PORT` | `chromadb.port` | `8000` |
| `MNEMONIC_CHROMADB_TOKEN` | `chromadb.token` | `my-token` |
| `MNEMONIC_SERVER_PORT` | `server.port` | `7438` |
| `DOLIBARR_URL` | `dolibarr.url` | `https://erp.example.com` |
| `DOLIBARR_API_KEY` | `dolibarr.api_key` | `your-api-key` |

### Verify configuration

```bash
mnemonic status
# Mnemonic Status
# ================
# ChromaDB:    http://localhost:8000
# Embeddings:  all-MiniLM-L6-v2 (384 dims)
# HTTP Server: 127.0.0.1:7438
# Dolibarr:    https://your-dolibarr.com
#
# Domains:
#   commercial      mn-commercial (6 types)
#   operations      mn-operations (6 types)
#   financial       mn-financial (6 types)
#   engineering     mn-engineering (6 types)
#   knowledge       mn-knowledge (5 types)
#   references      mn-references (3 types)
```

---

## Setting up ChromaDB

Mnemonic requires a ChromaDB server. You can run it with Docker:

```bash
docker run -d \
  --name chromadb \
  -p 8000:8000 \
  -v chromadb_data:/chroma/chroma \
  -e IS_PERSISTENT=TRUE \
  -e ANONYMIZED_TELEMETRY=FALSE \
  chromadb/chroma:latest
```

Or with Docker Compose:

```yaml
# docker-compose.yaml
services:
  chromadb:
    image: chromadb/chroma:latest
    ports:
      - "8000:8000"
    volumes:
      - chromadb_data:/chroma/chroma
    environment:
      IS_PERSISTENT: "TRUE"
      ANONYMIZED_TELEMETRY: "FALSE"

volumes:
  chromadb_data:
```

Verify ChromaDB is running:

```bash
curl http://localhost:8000/api/v2/heartbeat
# {"nanosecond heartbeat":...}
```

---

## Syncing from Dolibarr ERP

If you use [Dolibarr](https://www.dolibarr.org/) as your ERP, mnemonic can sync customers, projects, proposals, and products directly.

### First sync

```bash
# Full sync — imports everything
mnemonic sync-erp --full
```

### Incremental sync (default)

```bash
# Only changes since last sync (default: last 365 days)
mnemonic sync-erp

# Last 30 days only
mnemonic sync-erp --days=30
```

### Deep sync by client

```bash
# Sync everything for a specific client: projects, proposals, invoices
mnemonic sync-erp --client="Ecopetrol"
```

### Other sync options

```bash
# Sync only one entity type
mnemonic sync-erp --only=customers
mnemonic sync-erp --only=projects
mnemonic sync-erp --only=proposals
mnemonic sync-erp --only=products

# Preview what would be synced (no changes)
mnemonic sync-erp --dry-run

# Combine flags
mnemonic sync-erp --client="Ecopetrol" --days=90
```

### What gets synced

| Dolibarr | Mnemonic Domain | Entity Type |
|----------|----------------|-------------|
| Customers (thirdparties) | mn-commercial | client |
| Projects | mn-operations | project |
| Proposals | mn-commercial | proposal |
| Products/Services | mn-financial | apu / procurement |

---

## Claude Code Plugin

Mnemonic integrates with Claude Code as a plugin, providing:

- **MCP server** — 25 tools for searching, saving, and navigating the KB
- **Hooks** — proactive Knowledge Protocol injection, contextual search nudges
- **SKILL.md** — instructions for when to search and save automatically

### Install as plugin

```bash
# From the mnemonic repo
./scripts/install-plugin.sh
```

This:
1. Copies plugin files to `~/.claude/plugins/cache/mnemonic/`
2. Registers the plugin in Claude Code
3. Enables hooks for proactive behavior

**Restart Claude Code after installing.**

### Remove plugin

```bash
./scripts/install-plugin.sh --remove
```

### What the hooks do

| Hook | When | What it does |
|------|------|-------------|
| **SessionStart** | Opening Claude Code | Starts mnemonic serve, injects Knowledge Protocol, shows KB status |
| **UserPromptSubmit** | Each user message | First message: loads tools. Subsequent: contextual search nudges |
| **SubagentStop** | Sub-agent finishes | Captures knowledge from agent output (async) |
| **SessionStop** | Closing session | Cleanup, notifies server (async) |
| **Compaction** | Context compressed | Re-injects Knowledge Protocol |

### Using without the plugin

You can also use mnemonic as a standalone MCP server without the plugin hooks. Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "knowledge": {
      "command": "mnemonic",
      "args": ["mcp"],
      "env": {
        "MN_CONFIG": "./config/mnemonic.yaml"
      }
    }
  }
}
```

---

## Domains and Entity Types

Mnemonic organizes knowledge into 5 business domains plus a references domain:

### mn-commercial — Client and Sales

| Type | What it stores |
|------|---------------|
| `opportunity` | Detected business opportunity |
| `proposal` | Generated proposal/quote |
| `client` | Client profile and history |
| `competitor` | Competitor information |
| `client_comm` | Client communications (emails, calls) |
| `followup` | Commercial follow-up actions |

### mn-operations — Execution and Delivery

| Type | What it stores |
|------|---------------|
| `project` | Project in execution |
| `task` | Task or deliverable |
| `delivery` | Delivery milestone |
| `timeline` | Phase event |
| `quality` | Quality control checklist |
| `logistics` | Logistics and travel |

### mn-financial — Money

| Type | What it stores |
|------|---------------|
| `budget` | Project budget |
| `apu` | Unit price analysis |
| `procurement` | Purchase order |
| `invoice` | Invoice |
| `margin` | Margin analysis |
| `expense` | Expense or travel cost |

### mn-engineering — Technical Knowledge

| Type | What it stores |
|------|---------------|
| `architecture` | Architecture decision |
| `equipment` | Equipment selection rationale |
| `standard` | Applied standard or norm |
| `protocol` | Communication protocol |
| `config` | Technical configuration |
| `concept` | General technical knowledge |

### mn-knowledge — Lessons and Patterns

| Type | What it stores |
|------|---------------|
| `lesson` | Lesson learned |
| `decision` | Business decision with context |
| `conversation` | Work session summary |
| `agent_output` | Relevant agent output |
| `pattern` | Recurring pattern |

### mn-references — Relationships

| Type | What it stores |
|------|---------------|
| `reference` | PK-ID ↔ ERP code mapping |
| `relationship` | Entity-to-entity link |
| `sync_state` | Last sync timestamp |

---

## MCP Tools Reference

### Layer 0 — Inventory (no embeddings, instant)

| Tool | Description |
|------|-------------|
| `search_quick` | Fast metadata search without embeddings |
| `browse` | Paginated listing of a domain |
| `count` | Entity counts by domain |
| `list_types` | Available entity types |

### Layer 1 — Semantic Search (with embeddings)

| Tool | Description |
|------|-------------|
| `search` | Cross-domain semantic search |
| `search_commercial` | Search commercial domain |
| `search_operations` | Search operations domain |
| `search_financial` | Search financial domain |
| `search_engineering` | Search engineering domain |
| `search_knowledge` | Search knowledge domain |

### Layer 2 — Full Detail (on demand)

| Tool | Description |
|------|-------------|
| `get_entity` | Full document for one entity |
| `get_entities` | Batch get (max 10) |

### Layer 3 — Relationships

| Tool | Description |
|------|-------------|
| `find_related` | Connected entities |
| `link_entities` | Create relationship |
| `get_timeline` | Chronological history |

### Write

| Tool | Description |
|------|-------------|
| `save_entity` | Save/update entity with embeddings |
| `update_metadata` | Update metadata only |
| `create_reference` | Generate PK-ID |
| `link_erp_reference` | Link PK-ID to ERP codes |

### Delete / Refs / Admin

| Tool | Description |
|------|-------------|
| `delete_entity` | Remove entity |
| `get_reference` | Lookup by PK-ID |
| `search_references` | Search references |
| `knowledge_status` | KB status overview |

---

## Development

### Build

```bash
go build -o mnemonic ./cmd/mnemonic/
```

### Test

```bash
go test ./... -short
```

### Project structure

```
mnemonic/
├── cmd/mnemonic/          # CLI entry point + cobra commands
├── internal/
│   ├── chroma/            # ChromaDB client, filters, operations
│   ├── config/            # YAML config loader
│   ├── domains/           # Entity model, domain service, references
│   ├── embeddings/        # ONNX embedding engine (optional)
│   ├── http/              # HTTP server for hooks
│   ├── mcp/               # MCP server, tools, handlers
│   └── sync/              # Dolibarr REST client, sync engine
├── plugin/                # Claude Code plugin (hooks, scripts, SKILL.md)
├── scripts/               # Install scripts
└── config/                # Default config
```

---

## License

MIT
