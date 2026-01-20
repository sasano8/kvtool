/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd_load

import (
	"github.com/sasano8/kvtool/filesystems"
	"github.com/sasano8/kvtool/pkg/common"
	"github.com/sasano8/kvtool/pkg/decoders"
	"github.com/sasano8/kvtool/pkg/encoders"
	"github.com/spf13/cobra"
)

func init() {
}

// envCmd represents the env command
var CmdLoadFromDotenv = &cobra.Command{
	Use:   "dotenv",
	Short: "A brief description of your command",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var path string
		var result any
		var err error

		if len(args) == 0 {
			path = filesystems.StdinPath
		} else {
			path = args[0]
		}

		fs := filesystems.FsLocalFileCli{Path: path}

		result, err = fs.Load(decoders.DotenvToJson)
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
