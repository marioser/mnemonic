package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/marioser/mnemonic/internal/config"
	"github.com/marioser/mnemonic/internal/domains"
)

// SyncResult contains the results of a sync operation.
type SyncResult struct {
	Customers int           `json:"customers"`
	Projects  int           `json:"projects"`
	Proposals int           `json:"proposals"`
	Products  int           `json:"products"`
	Duration  time.Duration `json:"duration"`
	Errors    []string      `json:"errors,omitempty"`
}

// Engine orchestrates the sync from Dolibarr to the knowledge base.
type Engine struct {
	dol *DolibarrClient
	svc *domains.Service
	cfg *config.Config
}

// NewEngine creates a new sync engine.
func NewEngine(dol *DolibarrClient, svc *domains.Service, cfg *config.Config) *Engine {
	return &Engine{dol: dol, svc: svc, cfg: cfg}
}

// SyncOptions configures sync behavior.
type SyncOptions struct {
	Full       bool   // Full reimport (ignore delta)
	ClientName string // Deep sync for specific client
	ProjectRef string // Sync specific project
	OnlyEntity string // Only sync this entity type (customers, projects, proposals, products)
	Days       int    // Override delta days
	DryRun     bool   // Don't actually save, just report
}

// Run executes the sync with given options.
func (e *Engine) Run(ctx context.Context, opts SyncOptions) (*SyncResult, error) {
	start := time.Now()
	result := &SyncResult{}

	// Determine delta timestamp
	var since string
	if !opts.Full {
		days := e.cfg.Dolibarr.Sync.DeltaDays
		if opts.Days > 0 {
			days = opts.Days
		}
		since = time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	}

	// Deep sync by client
	if opts.ClientName != "" {
		return e.syncByClient(ctx, opts.ClientName, since, opts.DryRun)
	}

	batch := e.cfg.Dolibarr.Sync.BatchSize
	entities := e.cfg.Dolibarr.Sync.Entities

	// Sync customers
	if shouldSync(opts.OnlyEntity, "customers", entities) {
		n, errs := e.syncCustomers(ctx, since, batch, opts.DryRun)
		result.Customers = n
		result.Errors = append(result.Errors, errs...)
	}

	// Sync projects
	if shouldSync(opts.OnlyEntity, "projects", entities) {
		n, errs := e.syncProjects(ctx, since, "", batch, opts.DryRun)
		result.Projects = n
		result.Errors = append(result.Errors, errs...)
	}

	// Sync proposals
	if shouldSync(opts.OnlyEntity, "proposals", entities) {
		n, errs := e.syncProposals(ctx, since, "", batch, opts.DryRun)
		result.Proposals = n
		result.Errors = append(result.Errors, errs...)
	}

	// Sync products
	if shouldSync(opts.OnlyEntity, "products", entities) {
		n, errs := e.syncProducts(ctx, since, batch, opts.DryRun)
		result.Products = n
		result.Errors = append(result.Errors, errs...)
	}

	result.Duration = time.Since(start)
	return result, nil
}

func shouldSync(onlyEntity, entityType string, entities map[string]bool) bool {
	if onlyEntity != "" {
		return onlyEntity == entityType
	}
	enabled, ok := entities[entityType]
	return ok && enabled
}

func (e *Engine) syncCustomers(ctx context.Context, since string, limit int, dryRun bool) (int, []string) {
	customers, err := e.dol.FetchCustomers(since, limit)
	if err != nil {
		return 0, []string{fmt.Sprintf("fetch customers: %v", err)}
	}

	var errs []string
	count := 0
	for _, c := range customers {
		entity := &domains.Entity{
			ID:       fmt.Sprintf("erp-cli-%s", c.ID),
			Type:     "client",
			Domain:   "commercial",
			Title:    c.Name,
			Content:  buildCustomerContent(c),
			Source:   "erp-sync",
			Status:   "active",
			Extra: map[string]string{
				"erp_customer_id": c.ID,
			},
		}
		if c.Town != "" {
			entity.Extra["city"] = c.Town
		}

		if !dryRun {
			if _, err := e.svc.SaveEntity(ctx, entity); err != nil {
				errs = append(errs, fmt.Sprintf("save customer %s: %v", c.Name, err))
				continue
			}
		}
		count++
	}
	return count, errs
}

func (e *Engine) syncProjects(ctx context.Context, since, socID string, limit int, dryRun bool) (int, []string) {
	projects, err := e.dol.FetchProjects(since, socID, limit)
	if err != nil {
		return 0, []string{fmt.Sprintf("fetch projects: %v", err)}
	}

	var errs []string
	count := 0
	for _, p := range projects {
		entity := &domains.Entity{
			ID:      fmt.Sprintf("erp-proj-%s", p.ID),
			Type:    "project",
			Domain:  "operations",
			Title:   p.Title,
			Content: buildProjectContent(p),
			Source:  "erp-sync",
			Extra: map[string]string{
				"erp_project_ref": p.Ref,
			},
		}
		if p.SocID != "" {
			entity.ClientID = fmt.Sprintf("erp-cli-%s", p.SocID)
			entity.RelatedIDs = append(entity.RelatedIDs, entity.ClientID)
		}
		if p.Budget != "" {
			entity.Extra["budget_amount"] = p.Budget
		}

		if !dryRun {
			if _, err := e.svc.SaveEntity(ctx, entity); err != nil {
				errs = append(errs, fmt.Sprintf("save project %s: %v", p.Title, err))
				continue
			}
		}
		count++
	}
	return count, errs
}

func (e *Engine) syncProposals(ctx context.Context, since, socID string, limit int, dryRun bool) (int, []string) {
	proposals, err := e.dol.FetchProposals(since, socID, limit)
	if err != nil {
		return 0, []string{fmt.Sprintf("fetch proposals: %v", err)}
	}

	var errs []string
	count := 0
	for _, p := range proposals {
		entity := &domains.Entity{
			ID:      fmt.Sprintf("erp-prop-%s", p.ID),
			Type:    "proposal",
			Domain:  "commercial",
			Title:   fmt.Sprintf("Propuesta %s", p.Ref),
			Content: buildProposalContent(p),
			Source:  "erp-sync",
			Extra: map[string]string{
				"erp_proposal_ref": p.Ref,
				"total_ht":         p.TotalHT,
				"total_ttc":        p.TotalTTC,
			},
		}
		if p.SocID != "" {
			entity.ClientID = fmt.Sprintf("erp-cli-%s", p.SocID)
			entity.RelatedIDs = append(entity.RelatedIDs, entity.ClientID)
		}

		if !dryRun {
			if _, err := e.svc.SaveEntity(ctx, entity); err != nil {
				errs = append(errs, fmt.Sprintf("save proposal %s: %v", p.Ref, err))
				continue
			}
		}
		count++
	}
	return count, errs
}

func (e *Engine) syncProducts(ctx context.Context, since string, limit int, dryRun bool) (int, []string) {
	products, err := e.dol.FetchProducts(since, limit)
	if err != nil {
		return 0, []string{fmt.Sprintf("fetch products: %v", err)}
	}

	var errs []string
	count := 0
	for _, p := range products {
		typeName := "apu"
		if p.Type == "0" {
			typeName = "procurement" // physical product
		}

		entity := &domains.Entity{
			ID:      fmt.Sprintf("erp-prod-%s", p.ID),
			Type:    typeName,
			Domain:  "financial",
			Title:   p.Label,
			Content: buildProductContent(p),
			Source:  "erp-sync",
			Extra: map[string]string{
				"erp_ref": p.Ref,
				"price":   p.Price,
			},
		}

		if !dryRun {
			if _, err := e.svc.SaveEntity(ctx, entity); err != nil {
				errs = append(errs, fmt.Sprintf("save product %s: %v", p.Label, err))
				continue
			}
		}
		count++
	}
	return count, errs
}

func (e *Engine) syncByClient(ctx context.Context, clientName, since string, dryRun bool) (*SyncResult, error) {
	start := time.Now()
	result := &SyncResult{}

	// Find the client in Dolibarr
	customer, err := e.dol.FindCustomerByName(clientName)
	if err != nil {
		return nil, fmt.Errorf("finding client: %w", err)
	}

	socID := customer.ID
	batch := 500 // Higher limit for deep sync

	// Sync the client itself
	entity := &domains.Entity{
		ID:      fmt.Sprintf("erp-cli-%s", customer.ID),
		Type:    "client",
		Domain:  "commercial",
		Title:   customer.Name,
		Content: buildCustomerContent(*customer),
		Source:  "erp-sync",
		Status:  "active",
		Extra:   map[string]string{"erp_customer_id": customer.ID},
	}
	if !dryRun {
		e.svc.SaveEntity(ctx, entity)
	}
	result.Customers = 1

	// Sync projects for this client
	n, errs := e.syncProjects(ctx, since, socID, batch, dryRun)
	result.Projects = n
	result.Errors = append(result.Errors, errs...)

	// Sync proposals for this client
	n, errs = e.syncProposals(ctx, since, socID, batch, dryRun)
	result.Proposals = n
	result.Errors = append(result.Errors, errs...)

	result.Duration = time.Since(start)
	return result, nil
}

// Content builders

func buildCustomerContent(c Customer) string {
	parts := []string{fmt.Sprintf("Cliente: %s", c.Name)}
	if c.NameAlias != "" {
		parts = append(parts, fmt.Sprintf("Alias: %s", c.NameAlias))
	}
	if c.Town != "" {
		parts = append(parts, fmt.Sprintf("Ciudad: %s", c.Town))
	}
	if c.NotePublic != "" {
		parts = append(parts, StripHTML(c.NotePublic))
	}
	return strings.Join(parts, " | ")
}

func buildProjectContent(p Project) string {
	parts := []string{
		fmt.Sprintf("Proyecto: %s", p.Title),
		fmt.Sprintf("Ref: %s", p.Ref),
	}
	if desc := StripHTML(p.Description); desc != "" {
		parts = append(parts, desc)
	}
	if p.Budget != "" && p.Budget != "0" {
		parts = append(parts, fmt.Sprintf("Presupuesto: %s", p.Budget))
	}
	return strings.Join(parts, " | ")
}

func buildProposalContent(p Proposal) string {
	parts := []string{
		fmt.Sprintf("Propuesta: %s", p.Ref),
	}
	if p.TotalHT != "" {
		parts = append(parts, fmt.Sprintf("Total HT: %s", p.TotalHT))
	}
	if p.TotalTTC != "" {
		parts = append(parts, fmt.Sprintf("Total TTC: %s", p.TotalTTC))
	}
	return strings.Join(parts, " | ")
}

func buildProductContent(p Product) string {
	parts := []string{
		fmt.Sprintf("Producto: %s", p.Label),
		fmt.Sprintf("Ref: %s", p.Ref),
	}
	if desc := StripHTML(p.Description); desc != "" {
		parts = append(parts, desc)
	}
	if p.Price != "" {
		parts = append(parts, fmt.Sprintf("Precio: %s", p.Price))
	}
	return strings.Join(parts, " | ")
}
