---
name: knowledge
description: Mnemonic organizational knowledge base with semantic search. Use proactively when working with proposals, projects, clients, or engineering decisions.
---

# Mnemonic Knowledge Base

You have access to an organizational knowledge base with semantic search across 5 domains.

## The Onion Principle

Always use the **lightest layer first** to minimize token consumption:

| Layer | Tools | Tokens | When to use |
|-------|-------|--------|-------------|
| 0 | search_quick, browse, count | ~50/result | "What exists?" — scan first |
| 1 | search, search_* | ~200/result | "What's relevant?" — semantic match |
| 2 | get_entity | ~500-2000 | "Tell me everything" — only specific items |
| 3 | find_related | N × ~50 | "What's connected?" — graph traversal |

**NEVER** request Layer 2 for multiple entities at once. Start with Layer 0/1, then drill down.

## 5 Domains

- **commercial**: opportunities, proposals, clients, competitors, client communications, followups
- **operations**: projects, tasks, deliveries, timeline, quality, logistics
- **financial**: budgets, APU, procurement, invoices, margins, expenses
- **engineering**: architectures, equipment, standards, protocols, configs, concepts
- **knowledge**: lessons, decisions, conversations, agent outputs, patterns

## Proactive Search (do WITHOUT being asked)

- Before quoting → `search_commercial` + `search_engineering`
- Before estimating → `search_financial("APU similar")`
- Client mentioned → `search_quick(domain=commercial, client=X)`
- Technical decision → `search_engineering("previous architecture")`

## Proactive Save (do WITHOUT being asked)

- Technical decision made → `save_entity(domain=engineering, type=architecture)`
- Lesson learned → `save_entity(domain=knowledge, type=lesson)`
- Client email/communication → `save_entity(domain=commercial, type=client_comm)`
- Equipment selected → `save_entity(domain=engineering, type=equipment)`
- Proposal generated → `create_reference(ref_type=proposal)` + `save_entity`
