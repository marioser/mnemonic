package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/marioser/mnemonic/internal/chroma"
	"github.com/marioser/mnemonic/internal/config"
	"github.com/marioser/mnemonic/internal/domains"
	mnhttp "github.com/marioser/mnemonic/internal/http"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start HTTP server for hooks and admin",
	Long:  "Starts the Mnemonic HTTP server used by Claude Code hooks for health checks, context injection, and session management.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(projectDir)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Initialize ChromaDB client
		chromaClient, err := chroma.New(cfg)
		if err != nil {
			return fmt.Errorf("connecting to ChromaDB: %w", err)
		}
		defer chromaClient.Close()

		// Initialize service
		svc := domains.NewService(chromaClient, cfg)

		// Start HTTP server
		httpServer := mnhttp.New(cfg, svc)

		// Graceful shutdown
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			fmt.Fprintf(os.Stderr, "mnemonic: HTTP server starting on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
			if err := httpServer.Start(); err != nil && err.Error() != "http: Server closed" {
				fmt.Fprintf(os.Stderr, "mnemonic: server error: %v\n", err)
				os.Exit(1)
			}
		}()

		<-stop
		fmt.Fprintln(os.Stderr, "\nmnemonic: shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpServer.Shutdown(ctx)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
