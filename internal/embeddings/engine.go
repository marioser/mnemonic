package embeddings

import (
	"fmt"
	"sync"

	"github.com/amikos-tech/chroma-go/pkg/embeddings"
	defaultef "github.com/amikos-tech/chroma-go/pkg/embeddings/default_ef"

	"github.com/marioser/mnemonic/internal/config"
)

// Engine manages the embedding function lifecycle.
// Uses chroma-go's built-in ONNX runtime with all-MiniLM-L6-v2 model (384 dims).
// The model and ONNX runtime are auto-downloaded on first use to ~/.cache/chroma/.
type Engine struct {
	ef      *defaultef.DefaultEmbeddingFunction
	cleanup func() error
	cfg     *config.Config
	mu      sync.Mutex
	loaded  bool
}

// NewEngine creates a new embedding engine (lazy-loaded).
func NewEngine(cfg *config.Config) *Engine {
	return &Engine{cfg: cfg}
}

// Init initializes the ONNX runtime and loads the model.
// Called automatically on first use, but can be called explicitly for early initialization.
func (e *Engine) Init() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.loaded {
		return nil
	}

	ef, cleanup, err := defaultef.NewDefaultEmbeddingFunction()
	if err != nil {
		return fmt.Errorf("initializing embedding function: %w", err)
	}

	e.ef = ef
	e.cleanup = cleanup
	e.loaded = true
	return nil
}

// EmbeddingFunction returns the underlying chroma-go EmbeddingFunction interface.
// Initializes the engine if not already loaded.
func (e *Engine) EmbeddingFunction() (embeddings.EmbeddingFunction, error) {
	if err := e.Init(); err != nil {
		return nil, err
	}
	return e.ef, nil
}

// IsLoaded returns whether the model is loaded.
func (e *Engine) IsLoaded() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.loaded
}

// ModelInfo returns information about the loaded model.
func (e *Engine) ModelInfo() map[string]string {
	return map[string]string{
		"model":      "all-MiniLM-L6-v2",
		"dimensions": "384",
		"runtime":    "pure-onnx (Go, no CGO)",
		"cache_dir":  "~/.cache/chroma/",
	}
}

// Close releases the ONNX runtime resources.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.loaded {
		return nil
	}

	if e.ef != nil {
		e.ef.Close()
	}
	if e.cleanup != nil {
		if err := e.cleanup(); err != nil {
			return fmt.Errorf("cleaning up onnx runtime: %w", err)
		}
	}

	e.loaded = false
	return nil
}
