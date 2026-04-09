package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.ChromaDB.Host != "localhost" {
		t.Errorf("expected default host localhost, got %s", cfg.ChromaDB.Host)
	}
	if cfg.ChromaDB.Port != 8000 {
		t.Errorf("expected default port 8000, got %d", cfg.ChromaDB.Port)
	}
	if cfg.ChromaDB.CollectionPrefix != "mn" {
		t.Errorf("expected prefix mn, got %s", cfg.ChromaDB.CollectionPrefix)
	}
	if cfg.Server.Port != 7438 {
		t.Errorf("expected server port 7438, got %d", cfg.Server.Port)
	}
	if cfg.Embeddings.Dimensions != 768 {
		t.Errorf("expected 768 dims, got %d", cfg.Embeddings.Dimensions)
	}
	if cfg.Search.DefaultResults != 5 {
		t.Errorf("expected 5 default results, got %d", cfg.Search.DefaultResults)
	}
	if len(cfg.Domains) != 6 {
		t.Errorf("expected 6 domains, got %d", len(cfg.Domains))
	}
}

func TestCollectionName(t *testing.T) {
	cfg := Default()

	tests := []struct {
		domain   string
		expected string
	}{
		{"commercial", "mn-commercial"},
		{"operations", "mn-operations"},
		{"financial", "mn-financial"},
		{"engineering", "mn-engineering"},
		{"knowledge", "mn-knowledge"},
		{"references", "mn-references"},
		{"unknown", "mn-unknown"},
	}

	for _, tt := range tests {
		got := cfg.CollectionName(tt.domain)
		if got != tt.expected {
			t.Errorf("CollectionName(%s) = %s, want %s", tt.domain, got, tt.expected)
		}
	}
}

func TestValidDomain(t *testing.T) {
	cfg := Default()

	if !cfg.ValidDomain("commercial") {
		t.Error("commercial should be valid")
	}
	if !cfg.ValidDomain("engineering") {
		t.Error("engineering should be valid")
	}
	if cfg.ValidDomain("nonexistent") {
		t.Error("nonexistent should not be valid")
	}
}

func TestValidType(t *testing.T) {
	cfg := Default()

	if !cfg.ValidType("commercial", "proposal") {
		t.Error("commercial/proposal should be valid")
	}
	if !cfg.ValidType("engineering", "architecture") {
		t.Error("engineering/architecture should be valid")
	}
	if cfg.ValidType("commercial", "architecture") {
		t.Error("commercial/architecture should not be valid")
	}
	if cfg.ValidType("nonexistent", "anything") {
		t.Error("nonexistent domain should not be valid")
	}
}

func TestAllDomainNames(t *testing.T) {
	cfg := Default()
	names := cfg.AllDomainNames()

	if len(names) != 5 {
		t.Errorf("expected 5 domains (excluding references), got %d", len(names))
	}

	// references should not be in the list
	for _, n := range names {
		if n == "references" {
			t.Error("references should not be in AllDomainNames")
		}
	}
}

func TestChromaDBURL(t *testing.T) {
	cfg := Default()

	url := cfg.ChromaDBURL()
	if url != "http://localhost:8000" {
		t.Errorf("expected http://localhost:8000, got %s", url)
	}

	cfg.ChromaDB.SSL = true
	cfg.ChromaDB.Host = "chroma.example.com"
	cfg.ChromaDB.Port = 443
	url = cfg.ChromaDBURL()
	if url != "https://chroma.example.com:443" {
		t.Errorf("expected https://chroma.example.com:443, got %s", url)
	}
}

func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "config")
	os.MkdirAll(configDir, 0o755)

	yamlContent := `
chromadb:
  host: "192.168.1.100"
  port: 9000
  token: "test-token"
server:
  port: 8888
`
	os.WriteFile(filepath.Join(configDir, "mnemonic.yaml"), []byte(yamlContent), 0o644)

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ChromaDB.Host != "192.168.1.100" {
		t.Errorf("expected host 192.168.1.100, got %s", cfg.ChromaDB.Host)
	}
	if cfg.ChromaDB.Port != 9000 {
		t.Errorf("expected port 9000, got %d", cfg.ChromaDB.Port)
	}
	if cfg.ChromaDB.Token != "test-token" {
		t.Errorf("expected token test-token, got %s", cfg.ChromaDB.Token)
	}
	if cfg.Server.Port != 8888 {
		t.Errorf("expected server port 8888, got %d", cfg.Server.Port)
	}
	// Defaults should still be there for unset values
	if cfg.Embeddings.Dimensions != 768 {
		t.Errorf("expected default dimensions 768, got %d", cfg.Embeddings.Dimensions)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("MNEMONIC_CHROMADB_HOST", "env-host")
	t.Setenv("MNEMONIC_CHROMADB_PORT", "9999")
	t.Setenv("DOLIBARR_API_KEY", "my-secret-key")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.ChromaDB.Host != "env-host" {
		t.Errorf("expected env host, got %s", cfg.ChromaDB.Host)
	}
	if cfg.ChromaDB.Port != 9999 {
		t.Errorf("expected env port 9999, got %d", cfg.ChromaDB.Port)
	}
	if cfg.Dolibarr.APIKey != "my-secret-key" {
		t.Errorf("expected env api key, got %s", cfg.Dolibarr.APIKey)
	}
}

func TestLoadNonExistentDir(t *testing.T) {
	cfg, err := Load("/nonexistent/path")
	if err != nil {
		t.Fatalf("Load should not fail for missing dir: %v", err)
	}
	// Should return defaults
	if cfg.ChromaDB.Host != "localhost" {
		t.Errorf("expected default host, got %s", cfg.ChromaDB.Host)
	}
}
