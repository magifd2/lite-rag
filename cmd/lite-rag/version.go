package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at link time via -ldflags "-X main.version=<tag>".
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("lite-rag %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
