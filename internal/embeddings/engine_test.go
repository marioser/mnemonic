package embeddings

import (
	"testing"

	"github.com/marioser/mnemonic/internal/config"
)

func TestNewEngine(t *testing.T) {
	cfg := config.Default()
	engine := NewEngine(cfg)

	if engine == nil {
		t.Fatal("engine should not be nil")
	}
	if engine.IsLoaded() {
		t.Error("engine should not be loaded before Init")
	}
}

func TestModelInfo(t *testing.T) {
	cfg := config.Default()
	engine := NewEngine(cfg)

	info := engine.ModelInfo()
	if info["model"] != "all-MiniLM-L6-v2" {
		t.Errorf("expected all-MiniLM-L6-v2, got %s", info["model"])
	}
	if info["dimensions"] != "384" {
		t.Errorf("expected 384, got %s", info["dimensions"])
	}
	if info["runtime"] != "pure-onnx (Go, no CGO)" {
		t.Errorf("unexpected runtime: %s", info["runtime"])
	}
}

func TestCloseWithoutInit(t *testing.T) {
	cfg := config.Default()
	engine := NewEngine(cfg)

	// Close without init should not panic or error
	if err := engine.Close(); err != nil {
		t.Errorf("Close without Init should not error: %v", err)
	}
}

// Integration test — requires model download (~30MB first run)
// Run with: go test -run TestEngineInit -tags integration
func TestEngineInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg := config.Default()
	engine := NewEngine(cfg)
	defer engine.Close()

	if err := engine.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if !engine.IsLoaded() {
		t.Error("engine should be loaded after Init")
	}

	ef, err := engine.EmbeddingFunction()
	if err != nil {
		t.Fatalf("EmbeddingFunction failed: %v", err)
	}
	if ef == nil {
		t.Fatal("embedding function should not be nil")
	}
}
