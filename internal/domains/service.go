package domains

import (
	"context"
	"fmt"
	"time"

	chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"

	"github.com/sgsoluciones/mnemonic/internal/chroma"
	"github.com/sgsoluciones/mnemonic/internal/config"
)

// Service provides domain-level operations over the ChromaDB client.
type Service struct {
	client *chroma.Client
	cfg    *config.Config
}

// NewService creates a new domain service.
func NewService(client *chroma.Client, cfg *config.Config) *Service {
	return &Service{client: client, cfg: cfg}
}

// SaveEntity saves or updates an entity in the appropriate domain collection.
func (s *Service) SaveEntity(ctx context.Context, entity *Entity) (string, error) {
	if entity.Domain == "" {
		return "", fmt.Errorf("domain is required")
	}
	if entity.Type == "" {
		return "", fmt.Errorf("type is required")
	}
	if entity.Title == "" {
		return "", fmt.Errorf("title is required")
	}

	if !s.cfg.ValidDomain(entity.Domain) {
		return "", fmt.Errorf("invalid domain: %s", entity.Domain)
	}

	// Auto-generate ID if not provided
	if entity.ID == "" {
		entity.ID = fmt.Sprintf("%s-%s", entity.Type, time.Now().UTC().Format("20060102-150405"))
	}

	// Build document text for embedding
	doc := entity.BuildDocument()
	if doc == "" {
		return "", fmt.Errorf("entity produces empty document")
	}

	meta, err := entity.ToMetadata(s.cfg.Search.SummaryMaxChars)
	if err != nil {
		return "", fmt.Errorf("building metadata: %w", err)
	}

	if err := s.client.Upsert(ctx, entity.Domain, entity.ID, doc, meta); err != nil {
		return "", fmt.Errorf("upserting entity: %w", err)
	}

	return entity.ID, nil
}

// GetEntity retrieves a single entity by ID from a domain.
// If domain is empty, searches all domains.
func (s *Service) GetEntity(ctx context.Context, id string, domain string) (*Entity, error) {
	if domain != "" {
		return s.getFromDomain(ctx, id, domain)
	}

	// Search all domains
	for _, d := range s.cfg.AllDomainNames() {
		entity, err := s.getFromDomain(ctx, id, d)
		if err == nil && entity != nil {
			return entity, nil
		}
	}
	// Also check references
	entity, err := s.getFromDomain(ctx, id, "references")
	if err == nil && entity != nil {
		return entity, nil
	}

	return nil, fmt.Errorf("entity not found: %s", id)
}

func (s *Service) getFromDomain(ctx context.Context, id string, domain string) (*Entity, error) {
	result, err := s.client.GetByIDs(ctx, domain, []string{id}, true)
	if err != nil {
		return nil, err
	}
	ids := result.GetIDs()
	if len(ids) == 0 {
		return nil, nil
	}

	row := chromago.ResultRow{ID: ids[0]}
	docs := result.GetDocuments()
	if len(docs) > 0 {
		row.Document = docs[0].ContentString()
	}
	metas := result.GetMetadatas()
	if len(metas) > 0 {
		row.Metadata = metas[0]
	}

	entity := EntityFromResult(row, domain)
	return &entity, nil
}

// GetEntities retrieves multiple entities by IDs from a domain.
func (s *Service) GetEntities(ctx context.Context, ids []string, domain string) ([]Entity, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required for batch get")
	}

	result, err := s.client.GetByIDs(ctx, domain, ids, true)
	if err != nil {
		return nil, err
	}

	resultIDs := result.GetIDs()
	docs := result.GetDocuments()
	metas := result.GetMetadatas()

	var entities []Entity
	for i, id := range resultIDs {
		row := chromago.ResultRow{ID: id}
		if i < len(docs) {
			row.Document = docs[i].ContentString()
		}
		if i < len(metas) {
			row.Metadata = metas[i]
		}
		entities = append(entities, EntityFromResult(row, domain))
	}

	return entities, nil
}

// SearchQuick performs a metadata-only search (Layer 0, no embeddings).
func (s *Service) SearchQuick(ctx context.Context, domain string, filter chromago.WhereClause, limit int) ([]Entity, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required for search_quick")
	}
	if limit <= 0 {
		limit = s.cfg.Search.DefaultResults
	}

	result, err := s.client.GetByFilter(ctx, domain, filter, limit, 0, false)
	if err != nil {
		return nil, err
	}

	resultIDs := result.GetIDs()
	metas := result.GetMetadatas()

	var entities []Entity
	for i, id := range resultIDs {
		row := chromago.ResultRow{ID: id}
		if i < len(metas) {
			row.Metadata = metas[i]
		}
		entities = append(entities, EntityFromResult(row, domain))
	}

	return entities, nil
}

// Search performs semantic search (Layer 1, with embeddings).
func (s *Service) Search(ctx context.Context, query string, domain string, filter chromago.WhereClause, nResults int) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required for semantic search")
	}
	if nResults <= 0 {
		nResults = s.cfg.Search.DefaultResults
	}

	if domain != "" {
		return s.searchDomain(ctx, query, domain, filter, nResults)
	}

	// Cross-domain search
	results, err := s.client.QueryCrossDomain(ctx, query, filter, nResults)
	if err != nil {
		return nil, err
	}

	var searchResults []SearchResult
	for _, r := range results {
		entity := EntityFromResult(r.Row, r.Domain)
		searchResults = append(searchResults, SearchResult{
			Entity:     entity,
			Similarity: r.Similarity(),
		})
	}

	return searchResults, nil
}

func (s *Service) searchDomain(ctx context.Context, query string, domain string, filter chromago.WhereClause, nResults int) ([]SearchResult, error) {
	qr, err := s.client.QueryMetadataOnly(ctx, domain, query, filter, nResults)
	if err != nil {
		return nil, err
	}

	ids := qr.GetIDGroups()
	distances := qr.GetDistancesGroups()
	metas := qr.GetMetadatasGroups()

	var results []SearchResult
	if len(ids) > 0 {
		for i, id := range ids[0] {
			row := chromago.ResultRow{ID: id}
			if len(metas) > 0 && i < len(metas[0]) {
				row.Metadata = metas[0][i]
			}

			similarity := 0.0
			if len(distances) > 0 && i < len(distances[0]) {
				similarity = 1.0 - float64(distances[0][i])
			}

			entity := EntityFromResult(row, domain)
			results = append(results, SearchResult{
				Entity:     entity,
				Similarity: similarity,
			})
		}
	}

	return results, nil
}

// DeleteEntity removes an entity by ID from a domain.
func (s *Service) DeleteEntity(ctx context.Context, id string, domain string) error {
	if domain == "" {
		return fmt.Errorf("domain is required for delete")
	}
	return s.client.Delete(ctx, domain, []string{id})
}

// Browse lists entities in a domain with pagination (Layer 0).
func (s *Service) Browse(ctx context.Context, domain string, filter chromago.WhereClause, limit, offset int) ([]Entity, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain is required for browse")
	}
	if limit <= 0 {
		limit = 20
	}

	result, err := s.client.GetByFilter(ctx, domain, filter, limit, offset, false)
	if err != nil {
		return nil, err
	}

	resultIDs := result.GetIDs()
	metas := result.GetMetadatas()

	var entities []Entity
	for i, id := range resultIDs {
		row := chromago.ResultRow{ID: id}
		if i < len(metas) {
			row.Metadata = metas[i]
		}
		entities = append(entities, EntityFromResult(row, domain))
	}

	return entities, nil
}

// Count returns entity count for a domain, optionally filtered.
func (s *Service) Count(ctx context.Context, domain string) (int, error) {
	return s.client.CollectionCount(ctx, domain)
}

// AllCounts returns counts for all domains.
func (s *Service) AllCounts(ctx context.Context) (map[string]int, error) {
	return s.client.AllCounts(ctx)
}

// FindRelated finds entities related to a given entity ID (Layer 3).
func (s *Service) FindRelated(ctx context.Context, entityID string, domain string) ([]Entity, error) {
	// First get the entity to find its related_ids
	entity, err := s.GetEntity(ctx, entityID, domain)
	if err != nil {
		return nil, fmt.Errorf("finding entity: %w", err)
	}

	if len(entity.RelatedIDs) == 0 {
		return nil, nil
	}

	// Fetch related entities from all domains
	var related []Entity
	for _, d := range append(s.cfg.AllDomainNames(), "references") {
		result, err := s.client.GetByIDs(ctx, d, entity.RelatedIDs, false)
		if err != nil {
			continue
		}
		resultIDs := result.GetIDs()
		metas := result.GetMetadatas()
		for i, id := range resultIDs {
			row := chromago.ResultRow{ID: id}
			if i < len(metas) {
				row.Metadata = metas[i]
			}
			related = append(related, EntityFromResult(row, d))
		}
	}

	// Also find entities that reference this entity in their related_ids
	for _, d := range s.cfg.AllDomainNames() {
		filter := chroma.NewFilter().Eq("related_ids", entityID).Build()
		if filter == nil {
			continue
		}
		result, err := s.client.GetByFilter(ctx, d, filter, 10, 0, false)
		if err != nil {
			continue
		}
		resultIDs := result.GetIDs()
		metas := result.GetMetadatas()
		for i, id := range resultIDs {
			// Skip if already in related list
			found := false
			for _, r := range related {
				if r.ID == string(id) {
					found = true
					break
				}
			}
			if found {
				continue
			}
			row := chromago.ResultRow{ID: id}
			if i < len(metas) {
				row.Metadata = metas[i]
			}
			related = append(related, EntityFromResult(row, d))
		}
	}

	return related, nil
}

// LinkEntities creates a bidirectional relationship between two entities.
func (s *Service) LinkEntities(ctx context.Context, fromID, toID, relation string) error {
	// Save relationship in references collection
	relEntity := Entity{
		ID:         fmt.Sprintf("rel-%s-%s-%s", fromID, relation, toID),
		Type:       "relationship",
		Domain:     "references",
		Title:      fmt.Sprintf("%s %s %s", fromID, relation, toID),
		Content:    fmt.Sprintf("Relationship: %s -[%s]-> %s", fromID, relation, toID),
		RelatedIDs: []string{fromID, toID},
		Source:     "manual",
		Extra: map[string]string{
			"from_id":  fromID,
			"to_id":    toID,
			"relation": relation,
		},
	}

	_, err := s.SaveEntity(ctx, &relEntity)
	return err
}
