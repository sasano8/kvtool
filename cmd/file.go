/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
}

// CmdFile represents the file command (parent for load and convert)
var CmdFile = &cobra.Command{
	Use:   "file",
	Short: "Direct file operations (load, convert)",
	Long: `File operations for loading from various sources and converting between formats.

This is the low-level command for direct file access without using store configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}
