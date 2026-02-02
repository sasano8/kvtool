/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd_file_load

import (
	"github.com/sasano8/kvtool/pkg/common"
	"github.com/sasano8/kvtool/pkg/encoders"
	"github.com/sasano8/kvtool/pkg/filesystems"
	"github.com/spf13/cobra"
)

func init() {
}

// envCmd represents the env command
var CmdLoadFromEnv = &cobra.Command{
	Use:   "env",
	Short: "A brief description of your command",
	Long:  "",
	// SilenceUsage: true, // エラー時に Usage を表示しない。ただし、引数エラー時にも表示されなくなる
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		var result any
		var err error

		fs := filesystems.FsEnvFile{}
		result, err = fs.LoadAsJson()
		if err != nil {
			return err
		}

		out := filesystems.FsLocalFileCli{Path: "@", NoMkdir: false}
		w, err := out.OpenWriter()
		if err != nil {
			return err
		}
		defer w.Close()

		enc := &encoders.ObjToJsonEncoder{Newline: true}
		mars := common.NewMarshaller(enc, result)
		err = mars.Dump(w)
		return err
	},
}
