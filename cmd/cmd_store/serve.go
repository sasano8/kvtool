package cmd_store

import (
	"github.com/sasano8/kvtool/exceptions"
	"github.com/spf13/cobra"
)

func init() {
}

// envCmd represents the env command
var CmdServe = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long:  "",
	Args:  cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return exceptions.ErrSysNotImplemented
	},
}
