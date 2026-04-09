package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/marioser/mnemonic/internal/config"
)

// DolibarrClient is a direct REST client for Dolibarr API.
type DolibarrClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewDolibarrClient creates a new Dolibarr REST client.
func NewDolibarrClient(cfg *config.Config) (*DolibarrClient, error) {
	if cfg.Dolibarr.URL == "" {
		return nil, fmt.Errorf("dolibarr URL not configured")
	}
	if cfg.Dolibarr.APIKey == "" {
		return nil, fmt.Errorf("dolibarr API key not configured (set DOLIBARR_API_KEY)")
	}

	return &DolibarrClient{
		baseURL: strings.TrimRight(cfg.Dolibarr.URL, "/"),
		apiKey:  cfg.Dolibarr.APIKey,
		http: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

// get performs a GET request to the Dolibarr API.
func (c *DolibarrClient) get(endpoint string, params map[string]string) ([]byte, error) {
	u := fmt.Sprintf("%s/api/index.php%s", c.baseURL, endpoint)

	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("DOLAPIKEY", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dolibarr request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dolibarr API error %d: %s", resp.StatusCode, string(body[:min(200, len(body))]))
	}

	return body, nil
}

// Customer represents a Dolibarr third party.
type Customer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NameAlias   string `json:"name_alias"`
	Town        string `json:"town"`
	Client      string `json:"client"` // "1" = customer
	NotePublic  string `json:"note_public"`
	NotePrivate string `json:"note_private"`
	DateModify  string `json:"date_modification"`
}

// Project represents a Dolibarr project.
type Project struct {
	ID          string `json:"id"`
	Ref         string `json:"ref"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SocID       string `json:"socid"`
	Budget      string `json:"budget_amount"`
	DateStart   string `json:"date_start"`
	DateEnd     string `json:"date_end"`
	Status      string `json:"statut"`
	DateModify  string `json:"date_modification"`
}

// Proposal represents a Dolibarr commercial proposal.
type Proposal struct {
	ID         string `json:"id"`
	Ref        string `json:"ref"`
	RefClient  string `json:"ref_client"`
	SocID      string `json:"socid"`
	TotalHT    string `json:"total_ht"`
	TotalTTC   string `json:"total_ttc"`
	Status     string `json:"statut"`
	DateModify string `json:"date_modification"`
}

// Product represents a Dolibarr product/service.
type Product struct {
	ID          string `json:"id"`
	Ref         string `json:"ref"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Type        string `json:"type"` // "0"=product, "1"=service
	DateModify  string `json:"date_modification"`
}

// FetchCustomers fetches customers from Dolibarr with optional time filter.
func (c *DolibarrClient) FetchCustomers(since string, limit int) ([]Customer, error) {
	params := map[string]string{
		"sortfield": "t.tms",
		"sortorder": "DESC",
		"limit":     fmt.Sprintf("%d", limit),
	}
	if since != "" {
		params["sqlfilters"] = fmt.Sprintf("(t.tms:>:'%s')", since)
	}

	body, err := c.get("/thirdparties", params)
	if err != nil {
		return nil, err
	}

	var customers []Customer
	if err := json.Unmarshal(body, &customers); err != nil {
		return nil, fmt.Errorf("parsing customers: %w", err)
	}
	return customers, nil
}

// FetchProjects fetches projects with optional filters.
func (c *DolibarrClient) FetchProjects(since string, socID string, limit int) ([]Project, error) {
	params := map[string]string{
		"sortfield": "t.tms",
		"sortorder": "DESC",
		"limit":     fmt.Sprintf("%d", limit),
	}
	if since != "" {
		params["sqlfilters"] = fmt.Sprintf("(t.tms:>:'%s')", since)
	}
	if socID != "" {
		if params["sqlfilters"] != "" {
			params["sqlfilters"] = fmt.Sprintf("(t.tms:>:'%s') AND (t.fk_soc:=:%s)", since, socID)
		} else {
			params["sqlfilters"] = fmt.Sprintf("(t.fk_soc:=:%s)", socID)
		}
	}

	body, err := c.get("/projects", params)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, fmt.Errorf("parsing projects: %w", err)
	}
	return projects, nil
}

// FetchProposals fetches commercial proposals with optional filters.
func (c *DolibarrClient) FetchProposals(since string, socID string, limit int) ([]Proposal, error) {
	params := map[string]string{
		"sortfield": "t.tms",
		"sortorder": "DESC",
		"limit":     fmt.Sprintf("%d", limit),
	}
	if since != "" {
		params["sqlfilters"] = fmt.Sprintf("(t.tms:>:'%s')", since)
	}
	if socID != "" {
		if params["sqlfilters"] != "" {
			params["sqlfilters"] = fmt.Sprintf("(t.tms:>:'%s') AND (t.fk_soc:=:%s)", since, socID)
		} else {
			params["sqlfilters"] = fmt.Sprintf("(t.fk_soc:=:%s)", socID)
		}
	}

	body, err := c.get("/proposals", params)
	if err != nil {
		return nil, err
	}

	var proposals []Proposal
	if err := json.Unmarshal(body, &proposals); err != nil {
		return nil, fmt.Errorf("parsing proposals: %w", err)
	}
	return proposals, nil
}

// FetchProducts fetches products/services.
func (c *DolibarrClient) FetchProducts(since string, limit int) ([]Product, error) {
	params := map[string]string{
		"sortfield": "t.tms",
		"sortorder": "DESC",
		"limit":     fmt.Sprintf("%d", limit),
	}
	if since != "" {
		params["sqlfilters"] = fmt.Sprintf("(t.tms:>:'%s')", since)
	}

	body, err := c.get("/products", params)
	if err != nil {
		return nil, err
	}

	var products []Product
	if err := json.Unmarshal(body, &products); err != nil {
		return nil, fmt.Errorf("parsing products: %w", err)
	}
	return products, nil
}

// FindCustomerByName searches for a customer by name (partial match).
func (c *DolibarrClient) FindCustomerByName(name string) (*Customer, error) {
	params := map[string]string{
		"sqlfilters": fmt.Sprintf("(t.nom:like:'%%%s%%')", name),
		"limit":      "1",
	}

	body, err := c.get("/thirdparties", params)
	if err != nil {
		return nil, err
	}

	var customers []Customer
	if err := json.Unmarshal(body, &customers); err != nil {
		return nil, fmt.Errorf("parsing customer search: %w", err)
	}
	if len(customers) == 0 {
		return nil, fmt.Errorf("customer not found: %s", name)
	}
	return &customers[0], nil
}

var htmlRegex = regexp.MustCompile(`<[^>]+>`)

// StripHTML removes HTML tags and normalizes whitespace.
func StripHTML(text string) string {
	if text == "" {
		return ""
	}
	clean := htmlRegex.ReplaceAllString(text, " ")
	clean = regexp.MustCompile(`\s+`).ReplaceAllString(clean, " ")
	clean = strings.TrimSpace(clean)
	if len(clean) > 500 {
		return clean[:500]
	}
	return clean
}
