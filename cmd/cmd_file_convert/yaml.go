/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd_file_convert

import (
	"github.com/sasano8/kvtool/exceptions"
	"github.com/spf13/cobra"
)

func init() {
}

// envCmd represents the env command
var CmdConvertToYaml = &cobra.Command{
	Use:   "yaml",
	Short: "A brief description of your command",
	Long:  "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return exceptions.ErrSysNotImplemented
	},
}
