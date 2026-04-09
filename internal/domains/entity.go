package domains

import (
	"fmt"
	"strings"
	"time"

	chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"
)

// Entity represents a knowledge base entity across all domains.
type Entity struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Domain     string            `json:"domain"`
	Title      string            `json:"title"`
	Summary    string            `json:"summary,omitempty"`
	Content    string            `json:"content,omitempty"`
	ClientID   string            `json:"client_id,omitempty"`
	Industry   string            `json:"industry,omitempty"`
	Status     string            `json:"status,omitempty"`
	Source     string            `json:"source,omitempty"`
	RelatedIDs []string          `json:"related_ids,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Extra      map[string]string `json:"extra,omitempty"`
	CreatedAt  string            `json:"created_at,omitempty"`
	UpdatedAt  string            `json:"updated_at,omitempty"`
}

// BuildDocument constructs the text that will be embedded for semantic search.
// Concatenates key fields to maximize embedding quality.
func (e *Entity) BuildDocument() string {
	var parts []string

	if e.Title != "" {
		parts = append(parts, fmt.Sprintf("%s: %s", capitalize(e.Type), e.Title))
	}
	if e.ClientID != "" {
		parts = append(parts, fmt.Sprintf("Cliente: %s", e.ClientID))
	}
	if e.Industry != "" {
		parts = append(parts, fmt.Sprintf("Industria: %s", e.Industry))
	}
	if e.Content != "" {
		parts = append(parts, e.Content)
	}
	if len(e.Tags) > 0 {
		parts = append(parts, fmt.Sprintf("Tags: %s", strings.Join(e.Tags, ", ")))
	}

	return strings.Join(parts, " | ")
}

// BuildSummary returns the summary or a truncated version of content.
func (e *Entity) BuildSummary(maxChars int) string {
	if e.Summary != "" {
		if len(e.Summary) > maxChars {
			return e.Summary[:maxChars]
		}
		return e.Summary
	}
	if e.Content != "" {
		if len(e.Content) > maxChars {
			return e.Content[:maxChars]
		}
		return e.Content
	}
	return e.Title
}

// ToMetadata converts the entity to ChromaDB DocumentMetadata.
func (e *Entity) ToMetadata(summaryMaxChars int) (chromago.DocumentMetadata, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	attrs := []*chromago.MetaAttribute{
		chromago.NewStringAttribute("type", e.Type),
		chromago.NewStringAttribute("domain", e.Domain),
		chromago.NewStringAttribute("title", e.Title),
		chromago.NewStringAttribute("summary", e.BuildSummary(summaryMaxChars)),
		chromago.NewStringAttribute("updated_at", now),
	}

	if e.CreatedAt == "" {
		attrs = append(attrs, chromago.NewStringAttribute("created_at", now))
	} else {
		attrs = append(attrs, chromago.NewStringAttribute("created_at", e.CreatedAt))
	}
	if e.ClientID != "" {
		attrs = append(attrs, chromago.NewStringAttribute("client_id", e.ClientID))
	}
	if e.Industry != "" {
		attrs = append(attrs, chromago.NewStringAttribute("industry", e.Industry))
	}
	if e.Status != "" {
		attrs = append(attrs, chromago.NewStringAttribute("status", e.Status))
	}
	if e.Source != "" {
		attrs = append(attrs, chromago.NewStringAttribute("source", e.Source))
	}

	// Store related_ids as comma-separated string (ChromaDB metadata limitation for search)
	if len(e.RelatedIDs) > 0 {
		attrs = append(attrs, chromago.NewStringAttribute("related_ids", strings.Join(e.RelatedIDs, ",")))
	}
	if len(e.Tags) > 0 {
		attrs = append(attrs, chromago.NewStringAttribute("tags", strings.Join(e.Tags, ",")))
	}

	// Extra metadata
	for k, v := range e.Extra {
		attrs = append(attrs, chromago.NewStringAttribute(k, v))
	}

	return chromago.NewDocumentMetadata(attrs...), nil
}

// EntityFromResult reconstructs an Entity from a ChromaDB result row.
func EntityFromResult(row chromago.ResultRow, domain string) Entity {
	e := Entity{
		ID:     string(row.ID),
		Domain: domain,
	}

	if row.Metadata != nil {
		if v, ok := metaString(row.Metadata, "type"); ok {
			e.Type = v
		}
		if v, ok := metaString(row.Metadata, "domain"); ok {
			e.Domain = v
		}
		if v, ok := metaString(row.Metadata, "title"); ok {
			e.Title = v
		}
		if v, ok := metaString(row.Metadata, "summary"); ok {
			e.Summary = v
		}
		if v, ok := metaString(row.Metadata, "client_id"); ok {
			e.ClientID = v
		}
		if v, ok := metaString(row.Metadata, "industry"); ok {
			e.Industry = v
		}
		if v, ok := metaString(row.Metadata, "status"); ok {
			e.Status = v
		}
		if v, ok := metaString(row.Metadata, "source"); ok {
			e.Source = v
		}
		if v, ok := metaString(row.Metadata, "created_at"); ok {
			e.CreatedAt = v
		}
		if v, ok := metaString(row.Metadata, "updated_at"); ok {
			e.UpdatedAt = v
		}
		if v, ok := metaString(row.Metadata, "related_ids"); ok && v != "" {
			e.RelatedIDs = strings.Split(v, ",")
		}
		if v, ok := metaString(row.Metadata, "tags"); ok && v != "" {
			e.Tags = strings.Split(v, ",")
		}
	}

	if row.Document != "" {
		e.Content = row.Document
	}

	return e
}

// SearchResult wraps an entity with similarity score.
type SearchResult struct {
	Entity     Entity  `json:"entity"`
	Similarity float64 `json:"similarity"`
}

func metaString(meta chromago.DocumentMetadata, key string) (string, bool) {
	if meta == nil {
		return "", false
	}
	return meta.GetString(key)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
