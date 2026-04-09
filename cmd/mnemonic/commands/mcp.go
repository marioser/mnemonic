package commands

import (
	"fmt"
	"os"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/marioser/mnemonic/internal/chroma"
	"github.com/marioser/mnemonic/internal/config"
	"github.com/marioser/mnemonic/internal/domains"
	"github.com/marioser/mnemonic/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server (stdio transport for Claude Code)",
	Long:  "Starts the Mnemonic MCP server using stdio transport. This is typically called by Claude Code, not directly by users.",
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

		// Initialize services
		svc := domains.NewService(chromaClient, cfg)
		refSvc := domains.NewReferenceService(chromaClient, cfg, svc)

		// Create and serve MCP server
		s := mcp.NewServer(cfg, svc, refSvc)

		fmt.Fprintln(os.Stderr, "mnemonic: MCP server starting on stdio...")
		return mcpserver.ServeStdio(s)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
