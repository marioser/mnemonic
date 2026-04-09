package commands

import (
	"fmt"

	"github.com/marioser/mnemonic/internal/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show knowledge base status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(projectDir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		fmt.Println("Mnemonic Status")
		fmt.Println("================")
		fmt.Printf("ChromaDB:    %s\n", cfg.ChromaDBURL())
		fmt.Printf("Embeddings:  %s (%d dims)\n", cfg.Embeddings.Model, cfg.Embeddings.Dimensions)
		fmt.Printf("HTTP Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)

		if cfg.Dolibarr.URL != "" {
			fmt.Printf("Dolibarr:    %s\n", cfg.Dolibarr.URL)
		} else {
			fmt.Println("Dolibarr:    not configured")
		}

		fmt.Println("\nDomains:")
		for name, domain := range cfg.Domains {
			fmt.Printf("  %-15s %s (%d types)\n", name, domain.Collection, len(domain.Types))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
