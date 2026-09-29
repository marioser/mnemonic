package domains

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/marioser/mnemonic/internal/chroma"
	"github.com/marioser/mnemonic/internal/config"
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
	return rs.CreateReferenceForCOM(ctx, refType, name, client, "", erpRefs)
}

// CreateReferenceForCOM generates a PK-ID that carries the opportunity key.
func (rs *ReferenceService) CreateReferenceForCOM(ctx context.Context, refType, name, client, com string, erpRefs map[string]string) (string, error) {
	prefix, ok := rs.cfg.References.Types[refType]
	if !ok {
		return "", fmt.Errorf("unknown reference type: %s", refType)
	}

	// The sequential comes from the HIGHEST id of this type and year, not from
	// counting rows. Counting returned an id that was already issued as soon as
	// somebody deleted a reference, and the upsert overwrote it without a word.
	filter := chroma.NewFilter().Type("reference").Eq("ref_type", refType).Build()
	var existing []string
	if filter != nil {
		result, err := rs.client.GetByFilter(ctx, "references", filter, 0, 0, false)
		if err == nil {
			for _, id := range result.GetIDs() {
				existing = append(existing, string(id))
			}
		}
	}

	year := time.Now().UTC().Year()
	seq := nextSequential(existing, rs.cfg.References.Prefix, prefix, year)
	pkID := composePKID(rs.cfg.References.Prefix, prefix, year, seq, com)

	// An id already in use is never reused: the money, the files and the KB
	// history of another opportunity hang from it.
	for _, id := range existing {
		if id == pkID {
			return "", fmt.Errorf("%s is already registered: refusing to overwrite another opportunity's reference", pkID)
		}
	}

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

	if c := normalizeCOM(com); c != "" {
		entity.Extra["com"] = strings.Replace(c, "COM", "COM-", 1)
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
