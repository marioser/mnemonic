package chroma

import (
	chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"
)

// FilterBuilder helps construct ChromaDB WHERE clauses from optional parameters.
type FilterBuilder struct {
	clauses []chromago.WhereClause
}

// NewFilter creates a new FilterBuilder.
func NewFilter() *FilterBuilder {
	return &FilterBuilder{}
}

// Type adds a type filter if typeName is not empty.
func (fb *FilterBuilder) Type(typeName string) *FilterBuilder {
	if typeName != "" {
		fb.clauses = append(fb.clauses, chromago.EqString("type", typeName))
	}
	return fb
}

// Domain adds a domain filter if domain is not empty.
func (fb *FilterBuilder) Domain(domain string) *FilterBuilder {
	if domain != "" {
		fb.clauses = append(fb.clauses, chromago.EqString("domain", domain))
	}
	return fb
}

// Client adds a client_id filter if clientID is not empty.
func (fb *FilterBuilder) Client(clientID string) *FilterBuilder {
	if clientID != "" {
		fb.clauses = append(fb.clauses, chromago.EqString("client_id", clientID))
	}
	return fb
}

// Status adds a status filter if status is not empty.
func (fb *FilterBuilder) Status(status string) *FilterBuilder {
	if status != "" {
		fb.clauses = append(fb.clauses, chromago.EqString("status", status))
	}
	return fb
}

// Industry adds an industry filter if industry is not empty.
func (fb *FilterBuilder) Industry(industry string) *FilterBuilder {
	if industry != "" {
		fb.clauses = append(fb.clauses, chromago.EqString("industry", industry))
	}
	return fb
}

// Source adds a source filter if source is not empty.
func (fb *FilterBuilder) Source(source string) *FilterBuilder {
	if source != "" {
		fb.clauses = append(fb.clauses, chromago.EqString("source", source))
	}
	return fb
}

// RelatedTo adds a filter for entities related to a specific ID.
func (fb *FilterBuilder) RelatedTo(entityID string) *FilterBuilder {
	if entityID != "" {
		fb.clauses = append(fb.clauses, chromago.MetadataContainsString("related_ids", entityID))
	}
	return fb
}

// Eq adds an arbitrary string equality filter.
func (fb *FilterBuilder) Eq(key, value string) *FilterBuilder {
	if key != "" && value != "" {
		fb.clauses = append(fb.clauses, chromago.EqString(key, value))
	}
	return fb
}

// Build returns the composed WHERE clause.
// Returns nil if no filters were added (no filtering).
func (fb *FilterBuilder) Build() chromago.WhereClause {
	switch len(fb.clauses) {
	case 0:
		return nil
	case 1:
		return fb.clauses[0]
	default:
		return chromago.And(fb.clauses...)
	}
}

// HasFilters returns true if at least one filter was added.
func (fb *FilterBuilder) HasFilters() bool {
	return len(fb.clauses) > 0
}
