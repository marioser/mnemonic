package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// Layer 0 tools

func toolSearchQuick() mcp.Tool {
	return mcp.NewTool("search_quick",
		mcp.WithDescription("Fast metadata-only search without embeddings (Layer 0, ~50 tokens/result). Use this first to scan what exists before drilling deeper."),
		mcp.WithString("domain", mcp.Description("Domain to search: commercial, operations, financial, engineering, knowledge")),
		mcp.WithString("type", mcp.Description("Entity type filter (e.g. project, proposal, client, lesson)")),
		mcp.WithString("client", mcp.Description("Filter by client_id")),
		mcp.WithString("status", mcp.Description("Filter by status (active, completed, draft, won, lost)")),
		mcp.WithString("industry", mcp.Description("Filter by industry")),
		mcp.WithNumber("limit", mcp.Description("Max results (default 10)")),
	)
}

func toolBrowse() mcp.Tool {
	return mcp.NewTool("browse",
		mcp.WithDescription("List entities in a domain with pagination (Layer 0). For exploring what's in the KB."),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain to browse")),
		mcp.WithString("type", mcp.Description("Filter by entity type")),
		mcp.WithNumber("limit", mcp.Description("Max results (default 20)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
	)
}

func toolCount() mcp.Tool {
	return mcp.NewTool("count",
		mcp.WithDescription("Count entities by domain. Quick overview of KB size."),
		mcp.WithString("domain", mcp.Description("Specific domain to count (omit for all domains)")),
	)
}

func toolListTypes() mcp.Tool {
	return mcp.NewTool("list_types",
		mcp.WithDescription("List available entity types per domain. Useful for autocomplete."),
		mcp.WithString("domain", mcp.Description("Specific domain (omit for all)")),
	)
}

// Layer 1 tools

func toolSearch() mcp.Tool {
	return mcp.NewTool("search",
		mcp.WithDescription("Semantic search across domains using embeddings (Layer 1, ~200 tokens/result). Use for 'find similar', 'related to', conceptual queries. Without domain, searches all 5 domains in parallel."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Natural language search query")),
		mcp.WithString("domain", mcp.Description("Limit to specific domain")),
		mcp.WithString("type", mcp.Description("Filter by entity type")),
		mcp.WithString("client", mcp.Description("Filter by client_id")),
		mcp.WithString("industry", mcp.Description("Filter by industry")),
		mcp.WithString("status", mcp.Description("Filter by status")),
		mcp.WithNumber("n_results", mcp.Description("Max results (default 5)")),
	)
}

func toolSearchDomain(domain, description string) mcp.Tool {
	return mcp.NewTool("search_"+domain,
		mcp.WithDescription(description+" (Layer 1, semantic search with embeddings)"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Natural language search query")),
		mcp.WithString("type", mcp.Description("Filter by entity type")),
		mcp.WithString("client", mcp.Description("Filter by client_id")),
		mcp.WithString("industry", mcp.Description("Filter by industry")),
		mcp.WithString("status", mcp.Description("Filter by status")),
		mcp.WithNumber("n_results", mcp.Description("Max results (default 5)")),
	)
}

// Layer 2 tools

func toolGetEntity() mcp.Tool {
	return mcp.NewTool("get_entity",
		mcp.WithDescription("Get full document detail for a specific entity (Layer 2, ~500-2000 tokens). Only use when you need the complete content."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID")),
		mcp.WithString("domain", mcp.Description("Domain hint (faster if provided)")),
	)
}

func toolGetEntities() mcp.Tool {
	return mcp.NewTool("get_entities",
		mcp.WithDescription("Get full documents for multiple entities (Layer 2, max 10 IDs)."),
		mcp.WithString("ids", mcp.Required(), mcp.Description("Comma-separated entity IDs")),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain of the entities")),
	)
}

// Layer 3 tools

func toolFindRelated() mcp.Tool {
	return mcp.NewTool("find_related",
		mcp.WithDescription("Find entities related to a given entity (Layer 3, graph traversal). Returns related entities as Layer 0 results."),
		mcp.WithString("entity_id", mcp.Required(), mcp.Description("ID of the entity to find relations for")),
		mcp.WithString("domain", mcp.Description("Domain hint for the source entity")),
	)
}

func toolLinkEntities() mcp.Tool {
	return mcp.NewTool("link_entities",
		mcp.WithDescription("Create a relationship between two entities."),
		mcp.WithString("from_id", mcp.Required(), mcp.Description("Source entity ID")),
		mcp.WithString("to_id", mcp.Required(), mcp.Description("Target entity ID")),
		mcp.WithString("relation", mcp.Required(), mcp.Description("Relation type: originated_from, uses_service, decided_in, approved_by, delivered_to, paid_with, learned_from, depends_on, replaces")),
		mcp.WithString("context", mcp.Description("Why this relationship exists")),
	)
}

func toolGetTimeline() mcp.Tool {
	return mcp.NewTool("get_timeline",
		mcp.WithDescription("Get chronological history of an entity (proposal lifecycle, project phases)."),
		mcp.WithString("entity_id", mcp.Required(), mcp.Description("Entity ID (typically a proposal or project)")),
	)
}

// Write tools

func toolSaveEntity() mcp.Tool {
	return mcp.NewTool("save_entity",
		mcp.WithDescription("Save or update an entity in the knowledge base. Generates embeddings automatically."),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain: commercial, operations, financial, engineering, knowledge")),
		mcp.WithString("type", mcp.Required(), mcp.Description("Entity type (e.g. project, proposal, lesson, architecture)")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Short title for the entity")),
		mcp.WithString("content", mcp.Required(), mcp.Description("Full content text (will be embedded for semantic search)")),
		mcp.WithString("id", mcp.Description("Entity ID (auto-generated if not provided)")),
		mcp.WithString("summary", mcp.Description("Executive summary (auto-truncated from content if not provided)")),
		mcp.WithString("client_id", mcp.Description("Related client ID")),
		mcp.WithString("industry", mcp.Description("Industry sector")),
		mcp.WithString("status", mcp.Description("Entity status")),
		mcp.WithString("source", mcp.Description("Source: manual, erp-sync, agent, hook")),
		mcp.WithString("related_ids", mcp.Description("Comma-separated related entity IDs")),
		mcp.WithString("tags", mcp.Description("Comma-separated tags")),
	)
}

func toolUpdateMetadata() mcp.Tool {
	return mcp.NewTool("update_metadata",
		mcp.WithDescription("Update only metadata fields without regenerating embeddings. Use for status changes, adding tags, updating relations."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID")),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain of the entity")),
		mcp.WithString("status", mcp.Description("New status")),
		mcp.WithString("tags", mcp.Description("New comma-separated tags (replaces existing)")),
		mcp.WithString("related_ids", mcp.Description("New comma-separated related IDs (replaces existing)")),
	)
}

func toolCreateReference() mcp.Tool {
	return mcp.NewTool("create_reference",
		mcp.WithDescription("Create a new PK-ID reference (auto-generated sequential ID like PK-PROP-2026-0001)."),
		mcp.WithString("ref_type", mcp.Required(), mcp.Description("Type: proposal, project, client, decision, lesson, session")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Reference name")),
		mcp.WithString("client", mcp.Description("Client name or ID")),
		mcp.WithString("erp_proposal_ref", mcp.Description("Dolibarr proposal reference")),
		mcp.WithString("erp_project_ref", mcp.Description("Dolibarr project reference")),
		mcp.WithString("erp_customer_id", mcp.Description("Dolibarr customer ID")),
	)
}

func toolLinkERPReference() mcp.Tool {
	return mcp.NewTool("link_erp_reference",
		mcp.WithDescription("Link a PK-ID to ERP (Dolibarr) codes."),
		mcp.WithString("pk_id", mcp.Required(), mcp.Description("Internal PK-ID (e.g. PK-PROP-2026-0001)")),
		mcp.WithString("erp_proposal_ref", mcp.Description("Dolibarr proposal reference")),
		mcp.WithString("erp_project_ref", mcp.Description("Dolibarr project reference")),
		mcp.WithString("erp_order_ref", mcp.Description("Dolibarr order reference")),
		mcp.WithString("erp_invoice_ref", mcp.Description("Dolibarr invoice reference")),
		mcp.WithString("erp_customer_id", mcp.Description("Dolibarr customer ID")),
	)
}

// Delete tools

func toolDeleteEntity() mcp.Tool {
	return mcp.NewTool("delete_entity",
		mcp.WithDescription("Delete an entity from the knowledge base."),
		mcp.WithString("id", mcp.Required(), mcp.Description("Entity ID to delete")),
		mcp.WithString("domain", mcp.Required(), mcp.Description("Domain of the entity")),
	)
}

// Reference tools

func toolGetReference() mcp.Tool {
	return mcp.NewTool("get_reference",
		mcp.WithDescription("Get a reference by exact PK-ID lookup."),
		mcp.WithString("pk_id", mcp.Required(), mcp.Description("PK-ID (e.g. PK-PROP-2026-0001)")),
	)
}

func toolSearchReferences() mcp.Tool {
	return mcp.NewTool("search_references",
		mcp.WithDescription("Search references by PK-ID, ERP code, name, or client."),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query (PK-ID, ERP code, name, client)")),
		mcp.WithNumber("n_results", mcp.Description("Max results (default 5)")),
	)
}

// Admin tools

func toolKnowledgeStatus() mcp.Tool {
	return mcp.NewTool("knowledge_status",
		mcp.WithDescription("Get complete status of the knowledge base: collection counts, connection status, model info."),
	)
}
