package chroma

import (
	"context"
	"fmt"

	chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"
)

// GetByIDs retrieves entities by their IDs from a domain collection.
// Layer 2 operation — returns full documents + metadata.
func (c *Client) GetByIDs(ctx context.Context, domain string, ids []string, includeDoc bool) (chromago.GetResult, error) {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return nil, err
	}

	docIDs := toDocIDs(ids)
	includes := []chromago.Include{chromago.IncludeMetadatas}
	if includeDoc {
		includes = append(includes, chromago.IncludeDocuments)
	}

	return col.Get(ctx,
		chromago.WithIDs(docIDs...),
		chromago.WithInclude(includes...),
	)
}

// GetByFilter retrieves entities by metadata filter.
// Layer 0 operation — returns only metadata by default.
func (c *Client) GetByFilter(ctx context.Context, domain string, where chromago.WhereClause, limit, offset int, includeDoc bool) (chromago.GetResult, error) {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return nil, err
	}

	includes := []chromago.Include{chromago.IncludeMetadatas}
	if includeDoc {
		includes = append(includes, chromago.IncludeDocuments)
	}

	opts := []chromago.CollectionGetOption{
		chromago.WithInclude(includes...),
	}
	if where != nil {
		opts = append(opts, chromago.WithWhere(where))
	}
	if limit > 0 {
		opts = append(opts, chromago.WithLimit(limit))
	}
	if offset > 0 {
		opts = append(opts, chromago.WithOffset(offset))
	}

	return col.Get(ctx, opts...)
}

// Query performs semantic search in a domain collection.
// Layer 1 operation — uses embeddings, returns metadata + similarity.
func (c *Client) Query(ctx context.Context, domain string, queryText string, where chromago.WhereClause, nResults int) (chromago.QueryResult, error) {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return nil, err
	}

	if nResults <= 0 {
		nResults = c.cfg.Search.DefaultResults
	}

	opts := []chromago.CollectionQueryOption{
		chromago.WithQueryTexts(queryText),
		chromago.WithNResults(nResults),
		chromago.WithInclude(chromago.IncludeMetadatas, chromago.IncludeDocuments, chromago.IncludeDistances),
	}
	if where != nil {
		opts = append(opts, chromago.WithWhere(where))
	}

	return col.Query(ctx, opts...)
}

// QueryMetadataOnly performs semantic search returning only metadata (no documents).
// Optimized Layer 1 for token-conscious agents.
func (c *Client) QueryMetadataOnly(ctx context.Context, domain string, queryText string, where chromago.WhereClause, nResults int) (chromago.QueryResult, error) {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return nil, err
	}

	if nResults <= 0 {
		nResults = c.cfg.Search.DefaultResults
	}

	opts := []chromago.CollectionQueryOption{
		chromago.WithQueryTexts(queryText),
		chromago.WithNResults(nResults),
		chromago.WithInclude(chromago.IncludeMetadatas, chromago.IncludeDistances),
	}
	if where != nil {
		opts = append(opts, chromago.WithWhere(where))
	}

	return col.Query(ctx, opts...)
}

// QueryCrossDomain performs semantic search across all non-reference domains.
// Results are merged and sorted by similarity.
func (c *Client) QueryCrossDomain(ctx context.Context, queryText string, where chromago.WhereClause, nResults int) ([]ResultWithDomain, error) {
	domains := c.cfg.AllDomainNames()
	maxPerDomain := c.cfg.Search.MaxCrossDomainResults

	type domainResult struct {
		domain  string
		results []ResultWithDomain
		err     error
	}

	ch := make(chan domainResult, len(domains))
	for _, d := range domains {
		go func(domain string) {
			qr, err := c.QueryMetadataOnly(ctx, domain, queryText, where, maxPerDomain)
			if err != nil {
				ch <- domainResult{domain: domain, err: err}
				return
			}
			var rows []ResultWithDomain
			ids := qr.GetIDGroups()
			distances := qr.GetDistancesGroups()
			metadatas := qr.GetMetadatasGroups()
			if len(ids) > 0 {
				for i, id := range ids[0] {
					row := ResultWithDomain{
						Domain: domain,
						Row: chromago.ResultRow{
							ID: id,
						},
					}
					if len(distances) > 0 && i < len(distances[0]) {
						row.Row.Score = float64(distances[0][i])
					}
					if len(metadatas) > 0 && i < len(metadatas[0]) {
						row.Row.Metadata = metadatas[0][i]
					}
					rows = append(rows, row)
				}
			}
			ch <- domainResult{domain: domain, results: rows}
		}(d)
	}

	var all []ResultWithDomain
	for range domains {
		dr := <-ch
		if dr.err != nil {
			continue // skip failed domains
		}
		all = append(all, dr.results...)
	}

	// Sort by score descending (lower distance = better match)
	sortByScore(all)

	if nResults > 0 && len(all) > nResults {
		all = all[:nResults]
	}

	return all, nil
}

// Upsert inserts or updates an entity in a domain collection.
func (c *Client) Upsert(ctx context.Context, domain string, id string, document string, metadata chromago.DocumentMetadata) error {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return err
	}

	return col.Upsert(ctx,
		chromago.WithIDs(chromago.DocumentID(id)),
		chromago.WithTexts(document),
		chromago.WithMetadatas(metadata),
	)
}

// UpsertBatch inserts or updates multiple entities.
func (c *Client) UpsertBatch(ctx context.Context, domain string, ids []string, documents []string, metadatas []chromago.DocumentMetadata) error {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return err
	}

	if len(ids) != len(documents) || len(ids) != len(metadatas) {
		return fmt.Errorf("ids, documents, and metadatas must have the same length")
	}

	return col.Upsert(ctx,
		chromago.WithIDs(toDocIDs(ids)...),
		chromago.WithTexts(documents...),
		chromago.WithMetadatas(metadatas...),
	)
}

// Delete removes entities by IDs from a domain collection.
func (c *Client) Delete(ctx context.Context, domain string, ids []string) error {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return err
	}

	return col.Delete(ctx,
		chromago.WithIDs(toDocIDs(ids)...),
	)
}

// DeleteByFilter removes entities matching a metadata filter.
func (c *Client) DeleteByFilter(ctx context.Context, domain string, where chromago.WhereClause) error {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return err
	}

	return col.Delete(ctx,
		chromago.WithWhere(where),
	)
}

// ResultWithDomain pairs a query result row with its source domain.
type ResultWithDomain struct {
	Domain string
	Row    chromago.ResultRow
}

// Similarity returns 1 - distance (cosine distance → similarity score).
func (r ResultWithDomain) Similarity() float64 {
	return 1.0 - r.Row.Score
}

func toDocIDs(ids []string) []chromago.DocumentID {
	result := make([]chromago.DocumentID, len(ids))
	for i, id := range ids {
		result[i] = chromago.DocumentID(id)
	}
	return result
}

func sortByScore(results []ResultWithDomain) {
	// Sort by distance ascending (lower = better)
	for i := 1; i < len(results); i++ {
		key := results[i]
		j := i - 1
		for j >= 0 && results[j].Row.Score > key.Row.Score {
			results[j+1] = results[j]
			j--
		}
		results[j+1] = key
	}
}
