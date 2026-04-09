package mcp

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/sgsoluciones/mnemonic/internal/config"
	"github.com/sgsoluciones/mnemonic/internal/domains"
)

// NewServer creates a new MCP server with all knowledge tools registered.
func NewServer(cfg *config.Config, svc *domains.Service, refSvc *domains.ReferenceService) *server.MCPServer {
	s := server.NewMCPServer(
		"mnemonic",
		"0.1.0",
		server.WithToolCapabilities(true),
		server.WithInstructions(`Mnemonic — Organizational Knowledge Base with semantic search.
Use Layer 0 tools (search_quick, browse, count) for fast metadata lookups.
Use Layer 1 tools (search, search_*) for semantic search with embeddings.
Use Layer 2 tools (get_entity) only when you need full document detail.
Use Layer 3 tools (find_related) to explore entity relationships.
Always start with the lightest layer and drill down as needed.`),
	)

	h := &handlers{cfg: cfg, svc: svc, refSvc: refSvc}

	// Layer 0 — Inventory (no embeddings)
	s.AddTool(toolSearchQuick(), h.handleSearchQuick)
	s.AddTool(toolBrowse(), h.handleBrowse)
	s.AddTool(toolCount(), h.handleCount)
	s.AddTool(toolListTypes(), h.handleListTypes)

	// Layer 1 — Semantic search
	s.AddTool(toolSearch(), h.handleSearch)
	s.AddTool(toolSearchDomain("commercial", "Search commercial domain: opportunities, proposals, clients, competitors"), h.handleSearchCommercial)
	s.AddTool(toolSearchDomain("operations", "Search operations domain: projects, tasks, deliveries, timeline"), h.handleSearchOperations)
	s.AddTool(toolSearchDomain("financial", "Search financial domain: budgets, APU, procurement, invoices, margins"), h.handleSearchFinancial)
	s.AddTool(toolSearchDomain("engineering", "Search engineering domain: architectures, equipment, standards, protocols, configs"), h.handleSearchEngineering)
	s.AddTool(toolSearchDomain("knowledge", "Search knowledge domain: lessons, decisions, conversations, patterns"), h.handleSearchKnowledge)

	// Layer 2 — Detail
	s.AddTool(toolGetEntity(), h.handleGetEntity)
	s.AddTool(toolGetEntities(), h.handleGetEntities)

	// Layer 3 — Graph
	s.AddTool(toolFindRelated(), h.handleFindRelated)
	s.AddTool(toolLinkEntities(), h.handleLinkEntities)
	s.AddTool(toolGetTimeline(), h.handleGetTimeline)

	// Write
	s.AddTool(toolSaveEntity(), h.handleSaveEntity)
	s.AddTool(toolUpdateMetadata(), h.handleUpdateMetadata)
	s.AddTool(toolCreateReference(), h.handleCreateReference)
	s.AddTool(toolLinkERPReference(), h.handleLinkERPReference)

	// Delete
	s.AddTool(toolDeleteEntity(), h.handleDeleteEntity)

	// References
	s.AddTool(toolGetReference(), h.handleGetReference)
	s.AddTool(toolSearchReferences(), h.handleSearchReferences)

	// Admin
	s.AddTool(toolKnowledgeStatus(), h.handleKnowledgeStatus)

	return s
}
