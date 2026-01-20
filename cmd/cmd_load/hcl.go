/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd_load

import (
	"github.com/sasano8/kvtool/exceptions"
	"github.com/spf13/cobra"
)

func init() {
}

var CmdLoadFromHcl = &cobra.Command{
	Use:   "hcl",
	Short: "A brief description of your command",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exceptions.ErrSysNotImplemented
	},
}
