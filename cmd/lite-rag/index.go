package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"lite-rag/internal/config"
	"lite-rag/internal/database"
	"lite-rag/internal/indexer"
	"lite-rag/internal/llm"
)

var indexCmd = &cobra.Command{
	Use:   "index <directory>",
	Short: "Index Markdown documents in a directory",
	Args:  cobra.ExactArgs(1),
	RunE:  runIndex,
}

func init() {
	rootCmd.AddCommand(indexCmd)
}

func runIndex(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	client := llm.New(cfg.API.BaseURL, cfg.API.APIKey, cfg.Models.Embedding, cfg.Models.Chat)
	idx := indexer.New(db, client, cfg.Models.Embedding, cfg.Retrieval)

	dir := args[0]
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("directory %s: %w", dir, err)
	}

	fmt.Fprintf(os.Stderr, "Indexing %s...\n", dir)
	if err := idx.IndexDir(context.Background(), dir); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Done.")
	return nil
}
