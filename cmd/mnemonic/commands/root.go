package commands

import (
	"github.com/spf13/cobra"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"

	cfgFile    string
	projectDir string
)

var rootCmd = &cobra.Command{
	Use:   "mnemonic",
	Short: "Organizational knowledge management with semantic search",
	Long: `Mnemonic is an organizational knowledge management system that uses
semantic embeddings to store, search, and relate business knowledge
across commercial, operations, financial, engineering, and learning domains.

It integrates with Claude Code as a plugin and syncs data from Dolibarr ERP.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.mnemonic/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&projectDir, "project-dir", "", "project directory for local config")
}
