// Command lite-rag is the entry point for the lite-rag CLI.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"lite-rag/internal/config"
)

var (
	cfgFile    string
	dbOverride string
)

var rootCmd = &cobra.Command{
	Use:   "lite-rag",
	Short: "A CLI-based RAG document search system",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultConfigPath(), "config file path")
	rootCmd.PersistentFlags().StringVar(&dbOverride, "db", "", "database file path (overrides config database.path)")
}

// loadConfig loads the config file and applies any CLI-level overrides.
func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, err
	}
	if dbOverride != "" {
		cfg.Database.Path = dbOverride
	}
	return cfg, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
