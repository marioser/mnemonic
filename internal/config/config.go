package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ChromaDB   ChromaDBConfig   `yaml:"chromadb"`
	Embeddings EmbeddingsConfig `yaml:"embeddings"`
	Server     ServerConfig     `yaml:"server"`
	Search     SearchConfig     `yaml:"search"`
	Dolibarr   DolibarrConfig   `yaml:"dolibarr"`
	Domains    map[string]DomainConfig `yaml:"domains"`
	References ReferencesConfig `yaml:"references"`
	Log        LogConfig        `yaml:"log"`
}

type ChromaDBConfig struct {
	Host             string `yaml:"host"`
	Port             int    `yaml:"port"`
	Token            string `yaml:"token"`
	SSL              bool   `yaml:"ssl"`
	CollectionPrefix string `yaml:"collection_prefix"`
}

type EmbeddingsConfig struct {
	Model      string `yaml:"model"`
	Dimensions int    `yaml:"dimensions"`
	ModelPath  string `yaml:"model_path"`
	CacheEnabled bool `yaml:"cache_enabled"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type SearchConfig struct {
	DefaultResults        int     `yaml:"default_results"`
	MinSimilarity         float64 `yaml:"min_similarity"`
	MaxCrossDomainResults int     `yaml:"max_cross_domain_results"`
	SummaryMaxChars       int     `yaml:"summary_max_chars"`
}

type DolibarrConfig struct {
	URL    string           `yaml:"url"`
	APIKey string           `yaml:"api_key"`
	Sync   DolibarrSyncConfig `yaml:"sync"`
}

type DolibarrSyncConfig struct {
	DeltaDays int                    `yaml:"delta_days"`
	BatchSize int                    `yaml:"batch_size"`
	Entities  map[string]bool        `yaml:"entities"`
}

type DomainConfig struct {
	Collection string   `yaml:"collection"`
	Types      []string `yaml:"types"`
}

type ReferencesConfig struct {
	Prefix string            `yaml:"prefix"`
	Types  map[string]string `yaml:"types"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		ChromaDB: ChromaDBConfig{
			Host:             "localhost",
			Port:             8000,
			Token:            "",
			SSL:              false,
			CollectionPrefix: "mn",
		},
		Embeddings: EmbeddingsConfig{
			Model:        "all-MiniLM-L6-v2",
			Dimensions:   384,
			ModelPath:    filepath.Join(home, ".mnemonic", "models"),
			CacheEnabled: true,
		},
		Server: ServerConfig{
			Port: 7438,
			Host: "127.0.0.1",
		},
		Search: SearchConfig{
			DefaultResults:        5,
			MinSimilarity:         0.7,
			MaxCrossDomainResults: 15,
			SummaryMaxChars:       300,
		},
		Dolibarr: DolibarrConfig{
			URL:    "",
			APIKey: "",
			Sync: DolibarrSyncConfig{
				DeltaDays: 365,
				BatchSize: 50,
				Entities: map[string]bool{
					"customers": true,
					"projects":  true,
					"proposals": true,
					"products":  true,
					"invoices":  false,
					"orders":    false,
				},
			},
		},
		Domains: map[string]DomainConfig{
			"commercial": {
				Collection: "mn-commercial",
				Types:      []string{"opportunity", "proposal", "client", "competitor", "client_comm", "followup"},
			},
			"operations": {
				Collection: "mn-operations",
				Types:      []string{"project", "task", "delivery", "timeline", "quality", "logistics"},
			},
			"financial": {
				Collection: "mn-financial",
				Types:      []string{"budget", "apu", "procurement", "invoice", "margin", "expense"},
			},
			"engineering": {
				Collection: "mn-engineering",
				Types:      []string{"architecture", "equipment", "standard", "protocol", "config", "concept"},
			},
			"knowledge": {
				Collection: "mn-knowledge",
				Types:      []string{"lesson", "decision", "conversation", "agent_output", "pattern"},
			},
			"references": {
				Collection: "mn-references",
				Types:      []string{"reference", "relationship", "sync_state"},
			},
		},
		References: ReferencesConfig{
			Prefix: "PK",
			Types: map[string]string{
				"proposal": "PROP",
				"project":  "PROJ",
				"client":   "CLI",
				"decision": "DEC",
				"lesson":   "LES",
				"session":  "SES",
			},
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// Load reads configuration with priority:
// 1. CLI flags (applied after Load via ApplyOverrides)
// 2. Environment variables
// 3. Local config (projectDir/config/mnemonic.yaml)
// 4. Global config (~/.mnemonic/config.yaml)
// 5. Defaults
func Load(projectDir string) (*Config, error) {
	cfg := Default()

	// Global config
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, ".mnemonic", "config.yaml")
		if err := loadYAML(globalPath, cfg); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading global config: %w", err)
		}
	}

	// Local config
	if projectDir != "" {
		localPath := filepath.Join(projectDir, "config", "mnemonic.yaml")
		if err := loadYAML(localPath, cfg); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading local config: %w", err)
		}
	}

	// Environment variable config path override
	if envPath := os.Getenv("MN_CONFIG"); envPath != "" {
		if err := loadYAML(envPath, cfg); err != nil {
			return nil, fmt.Errorf("loading config from MN_CONFIG: %w", err)
		}
	}

	// Environment variable overrides
	applyEnvOverrides(cfg)

	return cfg, nil
}

func loadYAML(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("MNEMONIC_CHROMADB_HOST"); v != "" {
		cfg.ChromaDB.Host = v
	}
	if v := os.Getenv("MNEMONIC_CHROMADB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.ChromaDB.Port = port
		}
	}
	if v := os.Getenv("MNEMONIC_CHROMADB_TOKEN"); v != "" {
		cfg.ChromaDB.Token = v
	}
	if v := os.Getenv("MNEMONIC_SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("DOLIBARR_API_KEY"); v != "" {
		cfg.Dolibarr.APIKey = v
	}
	if v := os.Getenv("DOLIBARR_URL"); v != "" {
		cfg.Dolibarr.URL = v
	}
	if v := os.Getenv("MNEMONIC_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
}

// CollectionName returns the full collection name for a domain.
func (c *Config) CollectionName(domain string) string {
	if d, ok := c.Domains[domain]; ok {
		return d.Collection
	}
	return c.ChromaDB.CollectionPrefix + "-" + domain
}

// ValidDomain checks if a domain name is valid.
func (c *Config) ValidDomain(domain string) bool {
	_, ok := c.Domains[domain]
	return ok
}

// ValidType checks if a type is valid for a domain.
func (c *Config) ValidType(domain, typeName string) bool {
	d, ok := c.Domains[domain]
	if !ok {
		return false
	}
	for _, t := range d.Types {
		if t == typeName {
			return true
		}
	}
	return false
}

// AllDomainNames returns all domain names (excluding references).
func (c *Config) AllDomainNames() []string {
	var names []string
	for name := range c.Domains {
		if name != "references" {
			names = append(names, name)
		}
	}
	return names
}

// ChromaDBURL returns the full ChromaDB server URL.
func (c *Config) ChromaDBURL() string {
	scheme := "http"
	if c.ChromaDB.SSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, c.ChromaDB.Host, c.ChromaDB.Port)
}

// DataDir returns the mnemonic data directory.
func DataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mnemonic")
}

// Expand resolves ~ and env vars in a path.
func Expand(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}
