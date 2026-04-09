package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/marioser/mnemonic/internal/chroma"
	"github.com/marioser/mnemonic/internal/config"
	"github.com/marioser/mnemonic/internal/domains"
)

type handlers struct {
	cfg    *config.Config
	svc    *domains.Service
	refSvc *domains.ReferenceService
}

func textResult(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func jsonResult(v any) *mcp.CallToolResult {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("json marshal error: %v", err))
	}
	return mcp.NewToolResultText(string(data))
}

func errResult(msg string, args ...any) *mcp.CallToolResult {
	return mcp.NewToolResultError(fmt.Sprintf(msg, args...))
}

func paramStr(req mcp.CallToolRequest, key string) string {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func paramInt(req mcp.CallToolRequest, key string, defaultVal int) int {
	args := req.GetArguments()
	v, ok := args[key]
	if !ok {
		return defaultVal
	}
	f, ok := v.(float64)
	if !ok {
		return defaultVal
	}
	return int(f)
}

func buildFilter(req mcp.CallToolRequest) *chroma.FilterBuilder {
	return chroma.NewFilter().
		Type(paramStr(req, "type")).
		Client(paramStr(req, "client")).
		Status(paramStr(req, "status")).
		Industry(paramStr(req, "industry"))
}

// --- Layer 0 handlers ---

func (h *handlers) handleSearchQuick(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain := paramStr(req, "domain")
	if domain == "" {
		return errResult("domain is required for search_quick"), nil
	}

	filter := buildFilter(req).Build()
	limit := paramInt(req, "limit", 10)

	entities, err := h.svc.SearchQuick(ctx, domain, filter, limit)
	if err != nil {
		return errResult("search_quick failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"results": entities,
		"count":   len(entities),
		"layer":   0,
	}), nil
}

func (h *handlers) handleBrowse(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain := paramStr(req, "domain")
	if domain == "" {
		return errResult("domain is required"), nil
	}

	filter := chroma.NewFilter().Type(paramStr(req, "type")).Build()
	limit := paramInt(req, "limit", 20)
	offset := paramInt(req, "offset", 0)

	entities, err := h.svc.Browse(ctx, domain, filter, limit, offset)
	if err != nil {
		return errResult("browse failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"results": entities,
		"count":   len(entities),
		"domain":  domain,
		"layer":   0,
	}), nil
}

func (h *handlers) handleCount(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain := paramStr(req, "domain")

	if domain != "" {
		count, err := h.svc.Count(ctx, domain)
		if err != nil {
			return errResult("count failed: %v", err), nil
		}
		return jsonResult(map[string]any{"domain": domain, "count": count}), nil
	}

	counts, err := h.svc.AllCounts(ctx)
	if err != nil {
		return errResult("count failed: %v", err), nil
	}

	total := 0
	for _, c := range counts {
		if c > 0 {
			total += c
		}
	}

	return jsonResult(map[string]any{"counts": counts, "total": total}), nil
}

func (h *handlers) handleListTypes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain := paramStr(req, "domain")

	if domain != "" {
		d, ok := h.cfg.Domains[domain]
		if !ok {
			return errResult("unknown domain: %s", domain), nil
		}
		return jsonResult(map[string]any{domain: d.Types}), nil
	}

	result := make(map[string][]string)
	for name, d := range h.cfg.Domains {
		result[name] = d.Types
	}
	return jsonResult(result), nil
}

// --- Layer 1 handlers ---

func (h *handlers) handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := paramStr(req, "query")
	if query == "" {
		return errResult("query is required"), nil
	}

	domain := paramStr(req, "domain")
	filter := buildFilter(req).Build()
	nResults := paramInt(req, "n_results", 5)

	results, err := h.svc.Search(ctx, query, domain, filter, nResults)
	if err != nil {
		return errResult("search failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"results": results,
		"count":   len(results),
		"query":   query,
		"layer":   1,
	}), nil
}

func (h *handlers) searchByDomain(ctx context.Context, req mcp.CallToolRequest, domain string) (*mcp.CallToolResult, error) {
	query := paramStr(req, "query")
	if query == "" {
		return errResult("query is required"), nil
	}

	filter := buildFilter(req).Build()
	nResults := paramInt(req, "n_results", 5)

	results, err := h.svc.Search(ctx, query, domain, filter, nResults)
	if err != nil {
		return errResult("search_%s failed: %v", domain, err), nil
	}

	return jsonResult(map[string]any{
		"results": results,
		"count":   len(results),
		"query":   query,
		"domain":  domain,
		"layer":   1,
	}), nil
}

func (h *handlers) handleSearchCommercial(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.searchByDomain(ctx, req, "commercial")
}

func (h *handlers) handleSearchOperations(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.searchByDomain(ctx, req, "operations")
}

func (h *handlers) handleSearchFinancial(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.searchByDomain(ctx, req, "financial")
}

func (h *handlers) handleSearchEngineering(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.searchByDomain(ctx, req, "engineering")
}

func (h *handlers) handleSearchKnowledge(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.searchByDomain(ctx, req, "knowledge")
}

// --- Layer 2 handlers ---

func (h *handlers) handleGetEntity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := paramStr(req, "id")
	if id == "" {
		return errResult("id is required"), nil
	}

	entity, err := h.svc.GetEntity(ctx, id, paramStr(req, "domain"))
	if err != nil {
		return errResult("get_entity failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"entity": entity,
		"layer":  2,
	}), nil
}

func (h *handlers) handleGetEntities(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	idsStr := paramStr(req, "ids")
	domain := paramStr(req, "domain")
	if idsStr == "" || domain == "" {
		return errResult("ids and domain are required"), nil
	}

	ids := strings.Split(idsStr, ",")
	if len(ids) > 10 {
		return errResult("max 10 IDs per request"), nil
	}

	entities, err := h.svc.GetEntities(ctx, ids, domain)
	if err != nil {
		return errResult("get_entities failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"entities": entities,
		"count":    len(entities),
		"layer":    2,
	}), nil
}

// --- Layer 3 handlers ---

func (h *handlers) handleFindRelated(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	entityID := paramStr(req, "entity_id")
	if entityID == "" {
		return errResult("entity_id is required"), nil
	}

	related, err := h.svc.FindRelated(ctx, entityID, paramStr(req, "domain"))
	if err != nil {
		return errResult("find_related failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"entity_id": entityID,
		"related":   related,
		"count":     len(related),
		"layer":     3,
	}), nil
}

func (h *handlers) handleLinkEntities(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromID := paramStr(req, "from_id")
	toID := paramStr(req, "to_id")
	relation := paramStr(req, "relation")

	if fromID == "" || toID == "" || relation == "" {
		return errResult("from_id, to_id, and relation are required"), nil
	}

	if err := h.svc.LinkEntities(ctx, fromID, toID, relation); err != nil {
		return errResult("link_entities failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"status":   "linked",
		"from":     fromID,
		"to":       toID,
		"relation": relation,
	}), nil
}

func (h *handlers) handleGetTimeline(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	entityID := paramStr(req, "entity_id")
	if entityID == "" {
		return errResult("entity_id is required"), nil
	}

	// Timeline events are stored as type=timeline in operations domain
	filter := chroma.NewFilter().Type("timeline").Eq("related_ids", entityID).Build()
	entities, err := h.svc.SearchQuick(ctx, "operations", filter, 50)
	if err != nil {
		return errResult("get_timeline failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"entity_id": entityID,
		"events":    entities,
		"count":     len(entities),
		"layer":     3,
	}), nil
}

// --- Write handlers ---

func (h *handlers) handleSaveEntity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	entity := &domains.Entity{
		ID:       paramStr(req, "id"),
		Domain:   paramStr(req, "domain"),
		Type:     paramStr(req, "type"),
		Title:    paramStr(req, "title"),
		Content:  paramStr(req, "content"),
		Summary:  paramStr(req, "summary"),
		ClientID: paramStr(req, "client_id"),
		Industry: paramStr(req, "industry"),
		Status:   paramStr(req, "status"),
		Source:   paramStr(req, "source"),
	}

	if relIDs := paramStr(req, "related_ids"); relIDs != "" {
		entity.RelatedIDs = strings.Split(relIDs, ",")
	}
	if tags := paramStr(req, "tags"); tags != "" {
		entity.Tags = strings.Split(tags, ",")
	}

	id, err := h.svc.SaveEntity(ctx, entity)
	if err != nil {
		return errResult("save_entity failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"id":     id,
		"domain": entity.Domain,
		"type":   entity.Type,
		"status": "saved",
	}), nil
}

func (h *handlers) handleUpdateMetadata(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := paramStr(req, "id")
	domain := paramStr(req, "domain")
	if id == "" || domain == "" {
		return errResult("id and domain are required"), nil
	}

	// Get existing entity
	entity, err := h.svc.GetEntity(ctx, id, domain)
	if err != nil {
		return errResult("entity not found: %v", err), nil
	}

	// Update fields if provided
	if status := paramStr(req, "status"); status != "" {
		entity.Status = status
	}
	if tags := paramStr(req, "tags"); tags != "" {
		entity.Tags = strings.Split(tags, ",")
	}
	if relIDs := paramStr(req, "related_ids"); relIDs != "" {
		entity.RelatedIDs = strings.Split(relIDs, ",")
	}

	// Re-save (will regenerate embedding — acceptable tradeoff for simplicity)
	_, err = h.svc.SaveEntity(ctx, entity)
	if err != nil {
		return errResult("update failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"id":     id,
		"status": "updated",
	}), nil
}

func (h *handlers) handleCreateReference(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	refType := paramStr(req, "ref_type")
	name := paramStr(req, "name")
	if refType == "" || name == "" {
		return errResult("ref_type and name are required"), nil
	}

	erpRefs := make(map[string]string)
	if v := paramStr(req, "erp_proposal_ref"); v != "" {
		erpRefs["erp_proposal_ref"] = v
	}
	if v := paramStr(req, "erp_project_ref"); v != "" {
		erpRefs["erp_project_ref"] = v
	}
	if v := paramStr(req, "erp_customer_id"); v != "" {
		erpRefs["erp_customer_id"] = v
	}

	pkID, err := h.refSvc.CreateReference(ctx, refType, name, paramStr(req, "client"), erpRefs)
	if err != nil {
		return errResult("create_reference failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"pk_id":    pkID,
		"ref_type": refType,
		"name":     name,
	}), nil
}

func (h *handlers) handleLinkERPReference(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pkID := paramStr(req, "pk_id")
	if pkID == "" {
		return errResult("pk_id is required"), nil
	}

	erpRefs := make(map[string]string)
	for _, key := range []string{"erp_proposal_ref", "erp_project_ref", "erp_order_ref", "erp_invoice_ref", "erp_customer_id"} {
		if v := paramStr(req, key); v != "" {
			erpRefs[key] = v
		}
	}

	if len(erpRefs) == 0 {
		return errResult("at least one ERP reference is required"), nil
	}

	if err := h.refSvc.LinkERPReference(ctx, pkID, erpRefs); err != nil {
		return errResult("link_erp_reference failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"pk_id":  pkID,
		"linked": erpRefs,
	}), nil
}

// --- Delete handler ---

func (h *handlers) handleDeleteEntity(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := paramStr(req, "id")
	domain := paramStr(req, "domain")
	if id == "" || domain == "" {
		return errResult("id and domain are required"), nil
	}

	if err := h.svc.DeleteEntity(ctx, id, domain); err != nil {
		return errResult("delete failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"id":     id,
		"status": "deleted",
	}), nil
}

// --- Reference handlers ---

func (h *handlers) handleGetReference(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pkID := paramStr(req, "pk_id")
	if pkID == "" {
		return errResult("pk_id is required"), nil
	}

	entity, err := h.refSvc.GetReference(ctx, pkID)
	if err != nil {
		return errResult("reference not found: %v", err), nil
	}

	return jsonResult(map[string]any{
		"reference": entity,
	}), nil
}

func (h *handlers) handleSearchReferences(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := paramStr(req, "query")
	if query == "" {
		return errResult("query is required"), nil
	}

	nResults := paramInt(req, "n_results", 5)
	results, err := h.refSvc.SearchReferences(ctx, query, nResults)
	if err != nil {
		return errResult("search_references failed: %v", err), nil
	}

	return jsonResult(map[string]any{
		"results": results,
		"count":   len(results),
	}), nil
}

// --- Admin handler ---

func (h *handlers) handleKnowledgeStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	counts, err := h.svc.AllCounts(ctx)
	if err != nil {
		return errResult("status failed: %v", err), nil
	}

	total := 0
	for _, c := range counts {
		if c > 0 {
			total += c
		}
	}

	return jsonResult(map[string]any{
		"collections":    counts,
		"total_entities": total,
		"chromadb":       h.cfg.ChromaDBURL(),
		"embeddings": map[string]string{
			"model":      "all-MiniLM-L6-v2",
			"dimensions": "384",
			"runtime":    "pure-onnx",
		},
	}), nil
}
