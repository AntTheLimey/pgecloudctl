package cmd

import (
	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "pgecloudctl",
	Short: "CLI for managing pgEdge Cloud resources",
}

func Execute() error {
	return rootCmd.Execute()
}
