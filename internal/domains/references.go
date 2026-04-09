package domains

import (
	"context"
	"fmt"
	"time"

	"github.com/sgsoluciones/mnemonic/internal/chroma"
	"github.com/sgsoluciones/mnemonic/internal/config"
)

// ReferenceService manages PK-ID references and ERP linking.
type ReferenceService struct {
	client *chroma.Client
	cfg    *config.Config
	svc    *Service
}

// NewReferenceService creates a new reference service.
func NewReferenceService(client *chroma.Client, cfg *config.Config, svc *Service) *ReferenceService {
	return &ReferenceService{client: client, cfg: cfg, svc: svc}
}

// CreateReference generates a new PK-ID and registers it.
func (rs *ReferenceService) CreateReference(ctx context.Context, refType, name, client string, erpRefs map[string]string) (string, error) {
	prefix, ok := rs.cfg.References.Types[refType]
	if !ok {
		return "", fmt.Errorf("unknown reference type: %s", refType)
	}

	// Count existing references of this type to generate sequential ID
	filter := chroma.NewFilter().Type("reference").Eq("ref_type", refType).Build()
	var seq int
	if filter != nil {
		result, err := rs.client.GetByFilter(ctx, "references", filter, 0, 0, false)
		if err == nil {
			seq = len(result.GetIDs())
		}
	}
	seq++

	year := time.Now().UTC().Format("2006")
	pkID := fmt.Sprintf("%s-%s-%s-%04d", rs.cfg.References.Prefix, prefix, year, seq)

	entity := Entity{
		ID:     pkID,
		Type:   "reference",
		Domain: "references",
		Title:  fmt.Sprintf("%s: %s", pkID, name),
		Content: fmt.Sprintf("Reference %s | Type: %s | Name: %s | Client: %s",
			pkID, refType, name, client),
		ClientID: client,
		Source:   "manual",
		Extra: map[string]string{
			"ref_type":   refType,
			"pk_id":      pkID,
			"name":       name,
			"created_at": time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Add ERP references
	for k, v := range erpRefs {
		entity.Extra[k] = v
	}

	if _, err := rs.svc.SaveEntity(ctx, &entity); err != nil {
		return "", fmt.Errorf("saving reference: %w", err)
	}

	return pkID, nil
}

// GetReference retrieves a reference by PK-ID.
func (rs *ReferenceService) GetReference(ctx context.Context, pkID string) (*Entity, error) {
	return rs.svc.GetEntity(ctx, pkID, "references")
}

// LinkERPReference adds ERP codes to an existing PK-ID reference.
func (rs *ReferenceService) LinkERPReference(ctx context.Context, pkID string, erpRefs map[string]string) error {
	entity, err := rs.GetReference(ctx, pkID)
	if err != nil {
		return fmt.Errorf("reference not found: %s", pkID)
	}

	// Merge ERP refs into extra metadata
	if entity.Extra == nil {
		entity.Extra = make(map[string]string)
	}
	for k, v := range erpRefs {
		entity.Extra[k] = v
	}
	entity.Extra["linked_at"] = time.Now().UTC().Format(time.RFC3339)

	// Rebuild content with ERP codes for better embedding
	entity.Content = fmt.Sprintf("Reference %s | Name: %s | Client: %s",
		pkID, entity.Extra["name"], entity.ClientID)
	for k, v := range erpRefs {
		entity.Content += fmt.Sprintf(" | %s: %s", k, v)
	}

	_, err = rs.svc.SaveEntity(ctx, entity)
	return err
}

// SearchReferences searches across references by query text.
func (rs *ReferenceService) SearchReferences(ctx context.Context, query string, nResults int) ([]SearchResult, error) {
	return rs.svc.Search(ctx, query, "references", nil, nResults)
}
