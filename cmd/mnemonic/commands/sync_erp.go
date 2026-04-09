package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/marioser/mnemonic/internal/chroma"
	"github.com/marioser/mnemonic/internal/config"
	"github.com/marioser/mnemonic/internal/domains"
	mnSync "github.com/marioser/mnemonic/internal/sync"
)

var (
	syncFull    bool
	syncClient  string
	syncProject string
	syncOnly    string
	syncDays    int
	syncDryRun  bool
)

var syncERPCmd = &cobra.Command{
	Use:   "sync-erp",
	Short: "Sync data from Dolibarr ERP",
	Long: `Synchronize customers, projects, proposals, and products from Dolibarr ERP
into the Mnemonic knowledge base. Default: incremental sync (last year).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(projectDir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Initialize ChromaDB client (no embedding function needed)
		chromaClient, err := chroma.New(cfg)
		if err != nil {
			return fmt.Errorf("connecting to ChromaDB: %w", err)
		}
		defer chromaClient.Close()

		// Initialize Dolibarr client
		dolClient, err := mnSync.NewDolibarrClient(cfg)
		if err != nil {
			return fmt.Errorf("connecting to Dolibarr: %w", err)
		}

		// Initialize services
		svc := domains.NewService(chromaClient, cfg)
		syncEngine := mnSync.NewEngine(dolClient, svc, cfg)

		opts := mnSync.SyncOptions{
			Full:       syncFull,
			ClientName: syncClient,
			ProjectRef: syncProject,
			OnlyEntity: syncOnly,
			Days:       syncDays,
			DryRun:     syncDryRun,
		}

		if syncDryRun {
			fmt.Println("DRY RUN — no changes will be saved")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		fmt.Println("Syncing from Dolibarr...")
		result, err := syncEngine.Run(ctx, opts)
		if err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		fmt.Printf("\nSync complete in %s:\n", result.Duration.Round(time.Millisecond))
		fmt.Printf("  Customers:  %d\n", result.Customers)
		fmt.Printf("  Projects:   %d\n", result.Projects)
		fmt.Printf("  Proposals:  %d\n", result.Proposals)
		fmt.Printf("  Products:   %d\n", result.Products)

		if len(result.Errors) > 0 {
			fmt.Printf("\nErrors (%d):\n", len(result.Errors))
			for _, e := range result.Errors {
				fmt.Printf("  - %s\n", e)
			}
		}

		return nil
	},
}

func init() {
	syncERPCmd.Flags().BoolVar(&syncFull, "full", false, "Full reimport (delete and resync all)")
	syncERPCmd.Flags().StringVar(&syncClient, "client", "", "Deep sync for specific client (by name)")
	syncERPCmd.Flags().StringVar(&syncProject, "project", "", "Sync specific project (by ref)")
	syncERPCmd.Flags().StringVar(&syncOnly, "only", "", "Only sync entity type: customers, projects, proposals, products")
	syncERPCmd.Flags().IntVar(&syncDays, "days", 0, "Override delta days (default: from config)")
	syncERPCmd.Flags().BoolVar(&syncDryRun, "dry-run", false, "Show what would be synced without saving")

	rootCmd.AddCommand(syncERPCmd)
}
