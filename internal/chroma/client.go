package chroma

import (
	"context"
	"fmt"
	"sync"

	chromago "github.com/amikos-tech/chroma-go/pkg/api/v2"
	"github.com/amikos-tech/chroma-go/pkg/embeddings"

	"github.com/marioser/mnemonic/internal/config"
)

// Client wraps a ChromaDB HTTP client with collection management.
type Client struct {
	api         chromago.Client
	cfg         *config.Config
	collections map[string]chromago.Collection
	ef          embeddings.EmbeddingFunction
	mu          sync.RWMutex
}

// New creates a new ChromaDB client from config.
func New(cfg *config.Config, ef embeddings.EmbeddingFunction) (*Client, error) {
	opts := []chromago.ClientOption{
		chromago.WithBaseURL(cfg.ChromaDBURL()),
	}

	if cfg.ChromaDB.Token != "" {
		opts = append(opts, chromago.WithAuth(
			chromago.NewTokenAuthCredentialsProvider(
				cfg.ChromaDB.Token,
				chromago.AuthorizationTokenHeader,
			),
		))
	}

	api, err := chromago.NewHTTPClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("creating chromadb client: %w", err)
	}

	return &Client{
		api:         api,
		cfg:         cfg,
		collections: make(map[string]chromago.Collection),
		ef:          ef,
	}, nil
}

// Healthcheck verifies the ChromaDB server is reachable.
func (c *Client) Healthcheck(ctx context.Context) error {
	return c.api.Heartbeat(ctx)
}

// Close closes the underlying HTTP client.
func (c *Client) Close() error {
	return c.api.Close()
}

// Collection returns a collection by domain name, creating it if needed.
func (c *Client) Collection(ctx context.Context, domain string) (chromago.Collection, error) {
	name := c.cfg.CollectionName(domain)

	c.mu.RLock()
	col, ok := c.collections[name]
	c.mu.RUnlock()
	if ok {
		return col, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if col, ok := c.collections[name]; ok {
		return col, nil
	}

	col, err := c.api.GetOrCreateCollection(ctx, name,
		chromago.WithEmbeddingFunctionCreate(c.ef),
		chromago.WithHNSWSpaceCreate(embeddings.COSINE),
	)
	if err != nil {
		return nil, fmt.Errorf("getting collection %s: %w", name, err)
	}

	c.collections[name] = col
	return col, nil
}

// AllCollections returns all domain collections (excluding references).
func (c *Client) AllCollections(ctx context.Context) (map[string]chromago.Collection, error) {
	result := make(map[string]chromago.Collection)
	for _, domain := range c.cfg.AllDomainNames() {
		col, err := c.Collection(ctx, domain)
		if err != nil {
			return nil, err
		}
		result[domain] = col
	}
	return result, nil
}

// CollectionCount returns the document count for a collection.
func (c *Client) CollectionCount(ctx context.Context, domain string) (int, error) {
	col, err := c.Collection(ctx, domain)
	if err != nil {
		return 0, err
	}
	return col.Count(ctx)
}

// AllCounts returns document counts for all collections.
func (c *Client) AllCounts(ctx context.Context) (map[string]int, error) {
	counts := make(map[string]int)
	for domain := range c.cfg.Domains {
		count, err := c.CollectionCount(ctx, domain)
		if err != nil {
			counts[domain] = -1 // indicate error
			continue
		}
		counts[domain] = count
	}
	return counts, nil
}
