/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
}

// LoadCmd represents the load command
var CmdStore = &cobra.Command{
	Use:   "store",
	Short: "A brief description of your command",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
		cmd.Help()
	},
}
