// Command lite-rag is the entry point for the lite-rag CLI.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"lite-rag/internal/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "lite-rag",
	Short: "A CLI-based RAG document search system",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", config.DefaultConfigPath(), "config file path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
