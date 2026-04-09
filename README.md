# Mnemonic

Organizational knowledge management system with semantic search and progressive depth layers.

Built in Go. Designed for [Claude Code](https://docs.anthropic.com/en/docs/claude-code).

---

## Quick Start (5 minutes)

For users who want to get up and running fast:

```bash
# 1. Start ChromaDB (if you don't have one already)
docker run -d --name chromadb -p 8000:8000 \
  -v chromadb_data:/chroma/chroma \
  -e IS_PERSISTENT=TRUE -e ANONYMIZED_TELEMETRY=FALSE \
  chromadb/chroma:latest

# 2. Install mnemonic (detects OS/arch automatically, creates config)
curl -fsSL https://raw.githubusercontent.com/marioser/mnemonic/main/scripts/install.sh | bash

# 3. Edit config if ChromaDB is remote
#    nano ~/.mnemonic/config.yaml

# 4. Verify
mnemonic status

# 5. Install Claude Code plugin
claude plugin marketplace add marioser/mnemonic
claude plugin install mnemonic@mnemonic

# 6. Restart Claude Code — done!
```

**Alternative install methods:**

```bash
# Via Go (requires Go 1.24+ and GOPATH in PATH)
go install github.com/marioser/mnemonic/cmd/mnemonic@latest

# Manual download (replace OS and ARCH)
# OS: linux, darwin  |  ARCH: amd64, arm64
curl -L https://github.com/marioser/mnemonic/releases/latest/download/mnemonic_0.1.0_linux_amd64.tar.gz | tar xz
sudo mv mnemonic /usr/local/bin/
```

After restarting Claude Code, mnemonic will:
- Inject the **Knowledge Protocol** at session start
- Provide **25 MCP tools** for searching and saving knowledge
- Give **contextual nudges** when you mention clients, projects, or equipment

---

## Step-by-Step Guide for New Users

### Step 1: Set up ChromaDB

Mnemonic stores all knowledge in [ChromaDB](https://www.trychroma.com/), a vector database. You need a running ChromaDB server.

**Option A: Docker (easiest)**

```bash
docker run -d \
  --name chromadb \
  -p 8000:8000 \
  -v chromadb_data:/chroma/chroma \
  -e IS_PERSISTENT=TRUE \
  -e ANONYMIZED_TELEMETRY=FALSE \
  chromadb/chroma:latest
```

**Option B: Docker Compose**

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
    restart: unless-stopped

volumes:
  chromadb_data:
```

```bash
docker compose up -d
```

**Option C: Use an existing ChromaDB server**

If your team already has a ChromaDB server (e.g., on a remote server or cloud), just note its IP and port for the config step.

**Verify ChromaDB is running:**

```bash
curl http://localhost:8000/api/v2/heartbeat
# {"nanosecond heartbeat":...}
```

### Step 2: Install the mnemonic binary

You need the `mnemonic` binary in your system PATH.

**Option A: Go install (requires Go 1.24+)**

```bash
go install github.com/marioser/mnemonic/cmd/mnemonic@latest
```

**Option B: Download pre-built binary**

```bash
# macOS Apple Silicon (M1/M2/M3/M4)
curl -L https://github.com/marioser/mnemonic/releases/latest/download/mnemonic_darwin_arm64.tar.gz | tar xz
sudo cp mnemonic /usr/local/bin/

# macOS Intel
curl -L https://github.com/marioser/mnemonic/releases/latest/download/mnemonic_darwin_amd64.tar.gz | tar xz
sudo cp mnemonic /usr/local/bin/

# Linux x64
curl -L https://github.com/marioser/mnemonic/releases/latest/download/mnemonic_linux_amd64.tar.gz | tar xz
sudo cp mnemonic /usr/local/bin/

# Linux ARM
curl -L https://github.com/marioser/mnemonic/releases/latest/download/mnemonic_linux_arm64.tar.gz | tar xz
sudo cp mnemonic /usr/local/bin/
```

**Option C: Build from source**

```bash
git clone https://github.com/marioser/mnemonic.git
cd mnemonic
go build -o mnemonic ./cmd/mnemonic/
sudo cp mnemonic /usr/local/bin/
```

**Verify installation:**

```bash
mnemonic version
# mnemonic v0.1.0 (darwin/arm64)
```

### Step 3: Configure mnemonic

Create a global config file. Mnemonic will use this from any project.

```bash
mkdir -p ~/.mnemonic
cat > ~/.mnemonic/config.yaml << 'EOF'
# ChromaDB connection
chromadb:
  host: "localhost"        # Change to your ChromaDB IP if remote
  port: 8000
  token: ""                # Bearer token if auth is enabled
  ssl: false
  collection_prefix: "mn"

# HTTP server (used by hooks)
server:
  port: 7438
  host: "127.0.0.1"

# Search defaults
search:
  default_results: 5
  min_similarity: 0.7

# Dolibarr ERP (optional — skip if you don't use Dolibarr)
dolibarr:
  url: ""                  # e.g. "https://your-dolibarr.com"
  api_key: ""              # Your DOLAPIKEY from Dolibarr user settings
  sync:
    batch_size: 100
    entities:
      customers: true
      projects: true
      proposals: true
      products: true

log:
  level: "info"
EOF
```

**For project-specific config**, create `config/mnemonic.yaml` in your project directory. This overrides the global config for that project.

**Verify config:**

```bash
mnemonic status
```

You should see:

```
Mnemonic Status
================
ChromaDB:    http://localhost:8000
Embeddings:  all-MiniLM-L6-v2 (384 dims)
HTTP Server: 127.0.0.1:7438
Dolibarr:    not configured
ONNX Model:  not downloaded (run: mnemonic model download)

Domains:
  commercial      mn-commercial (6 types)
  operations      mn-operations (6 types)
  financial       mn-financial (6 types)
  engineering     mn-engineering (6 types)
  knowledge       mn-knowledge (5 types)
  references      mn-references (3 types)
```

### Step 4: Install the Claude Code plugin

This step installs mnemonic as a plugin with hooks that make it proactive (auto-inject Knowledge Protocol, contextual search nudges, agent output capture).

```bash
# Register the mnemonic marketplace
claude plugin marketplace add marioser/mnemonic

# Install the plugin
claude plugin install mnemonic@mnemonic
```

Output:

```
Adding marketplace...
✔ Successfully added marketplace: mnemonic

Installing plugin "mnemonic@mnemonic"...
✔ Successfully installed plugin: mnemonic@mnemonic (scope: user)
```

**Restart Claude Code** after installing.

### Step 5: Verify everything works

After restarting Claude Code:

1. The **Knowledge Protocol** should appear in the session context (you'll see "Mnemonic Knowledge Base — ACTIVE" at the top)
2. Check the KB status from Claude Code — ask Claude to run `knowledge_status`
3. Try a search — ask Claude "search for projects about automation"

### Step 6 (Optional): Sync from Dolibarr ERP

If you use Dolibarr, populate the KB with your existing data:

```bash
# Set your config with Dolibarr URL and API key, then:
MN_CONFIG=~/.mnemonic/config.yaml mnemonic sync-erp --full
```

### Step 7 (Optional): Project-specific config

For a specific project, create a local config that overrides the global one:

```bash
mkdir -p config
cat > config/mnemonic.yaml << 'EOF'
chromadb:
  host: "192.168.1.100"    # Your team's shared ChromaDB
  port: 8000
  token: "your-team-token"
EOF
```

Then add to your project's `.mcp.json`:

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
Layer 0 — Inventory (~50 tokens/result)
  "What exists?" — IDs, titles, metadata. No embeddings. Instant.

Layer 1 — Discovery (~200 tokens/result)
  "What's relevant?" — Semantic search + summary. Uses embeddings.

Layer 2 — Detail (~500-2000 tokens)
  "Tell me everything" — Full document + metadata. On demand only.

Layer 3 — Context (~N x 50 tokens)
  "What's connected?" — Graph traversal. Related entities, timeline.
```

An agent searching for "similar SCADA projects" gets Layer 0+1 results. If it needs more detail on a specific project, it requests Layer 2 for that one entity. **Never 2000 tokens for 10 projects when 50 tokens each was enough.**

---

## Architecture

```
mnemonic (Go binary)
  |
  |-- MCP Server (stdio, 25 tools) ----> Claude Code
  |
  |-- HTTP Server (:7438) ----> Hooks (session-start, user-prompt, etc.)
  |
  |-- ChromaDB Client ----> ChromaDB Server (remote, 6 collections)
  |
  |-- Dolibarr Client ----> Dolibarr ERP (REST API, delta sync)
```

---

## Configuration Reference

Mnemonic looks for configuration in this priority order:

1. `MN_CONFIG` environment variable (highest priority)
2. `./config/mnemonic.yaml` (project-local)
3. `~/.mnemonic/config.yaml` (global)
4. Built-in defaults

### Environment variables

Any config value can be overridden:

| Variable | Config path | Example |
|----------|------------|---------|
| `MN_CONFIG` | — | Path to config file |
| `MNEMONIC_CHROMADB_HOST` | `chromadb.host` | `192.168.1.100` |
| `MNEMONIC_CHROMADB_PORT` | `chromadb.port` | `8000` |
| `MNEMONIC_CHROMADB_TOKEN` | `chromadb.token` | `my-token` |
| `MNEMONIC_SERVER_PORT` | `server.port` | `7438` |
| `DOLIBARR_URL` | `dolibarr.url` | `https://erp.example.com` |
| `DOLIBARR_API_KEY` | `dolibarr.api_key` | `your-api-key` |

---

## Syncing from Dolibarr ERP

If you use [Dolibarr](https://www.dolibarr.org/) as your ERP, mnemonic can sync customers, projects, proposals, and products directly via REST API.

```bash
# Full sync (first time)
mnemonic sync-erp --full

# Incremental sync (default — last 365 days of changes)
mnemonic sync-erp

# Deep sync one client (all their projects, proposals, invoices)
mnemonic sync-erp --client="Ecopetrol"

# Only one entity type
mnemonic sync-erp --only=customers

# Last 30 days
mnemonic sync-erp --days=30

# Preview without saving
mnemonic sync-erp --dry-run
```

| Dolibarr | Mnemonic Domain | Entity Type |
|----------|----------------|-------------|
| Customers (thirdparties) | mn-commercial | client |
| Projects | mn-operations | project |
| Proposals | mn-commercial | proposal |
| Products/Services | mn-financial | apu / procurement |

---

## Claude Code Plugin

### What the hooks do

| Hook | When | What it does |
|------|------|-------------|
| **SessionStart** | Opening Claude Code | Starts `mnemonic serve`, injects Knowledge Protocol, shows KB status |
| **UserPromptSubmit** | Each user message | First message: loads MCP tools. Subsequent: contextual search nudges |
| **SubagentStop** | Sub-agent finishes | Captures knowledge from agent output (async) |
| **SessionStop** | Closing session | Cleanup and notifies server (async) |
| **Compaction** | Context compressed | Re-injects Knowledge Protocol to recover context |

### Install / Remove

```bash
# Add marketplace (one time)
claude plugin marketplace add marioser/mnemonic

# Install
claude plugin install mnemonic@mnemonic

# Remove
claude plugin uninstall mnemonic@mnemonic
```

### Using without the plugin (MCP only)

If you don't want hooks, just add to your `.mcp.json`:

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

This gives you the 25 tools without proactive behavior.

---

## Domains and Entity Types

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
| `reference` | PK-ID to ERP code mapping |
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

## Troubleshooting

### "mnemonic: command not found"

The binary is not in your PATH. After building or downloading, copy it:

```bash
sudo cp mnemonic /usr/local/bin/
# or on macOS with Homebrew:
cp mnemonic /opt/homebrew/bin/
```

### ChromaDB connection refused

Check that ChromaDB is running and accessible:

```bash
curl http://localhost:8000/api/v2/heartbeat
```

If using a remote server, verify the host/port in your config.

### Sync takes too long or times out

Reduce the batch size in your config:

```yaml
dolibarr:
  sync:
    batch_size: 50  # Lower from 100
```

Or sync one entity type at a time:

```bash
mnemonic sync-erp --only=customers
mnemonic sync-erp --only=projects
```

### Hooks not working after install

1. Make sure you restarted Claude Code
2. Verify plugin files exist: `ls ~/.claude/plugins/cache/mnemonic/mnemonic/0.1.0/`
3. Check that `mnemonic` binary is in PATH
4. Check settings: `cat ~/.claude/settings.json | grep mnemonic`

### Plugin conflicts with existing knowledge MCP

If your project `.mcp.json` already has a `knowledge` server configured, the plugin's MCP and the project MCP may conflict. Remove one:

- **Keep project MCP**: Remove plugin with `./scripts/install-plugin.sh --remove`
- **Keep plugin MCP**: Remove the `knowledge` entry from your `.mcp.json`

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

## Author

**Mario Serrano** — [github.com/marioser](https://github.com/marioser)

## License

MIT
