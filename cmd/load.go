/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
}

// CmdLoad represents the load command
var CmdLoad = &cobra.Command{
	Use:   "load",
	Short: "A brief description of your command",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		CmdRoot.Flags().BoolP("toggle", "t", false, "Help message for toggle")
		cmd.Help()
	},
}
