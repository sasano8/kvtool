package cmd_store

import (
	"fmt"

	"github.com/sasano8/kvtool/filesystems"
	"github.com/sasano8/kvtool/pkg/common"
	"github.com/sasano8/kvtool/pkg/encoders"
	"github.com/spf13/cobra"
)

func init() {
}

type Store struct {
	Type string         `json:"type"`
	Args map[string]any `json:"args"`
}

type StoreConfig struct {
	Version    float64                     `json:"version"`
	Context    string                      `json:"context"`
	Namespaces map[string]map[string]Store `json:"namespaces"`
}

// envCmd represents the env command
var CmdInit = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long:  "",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		force := *cmd.Flags().BoolP("force", "f", false, "Help message for toggle")
		path := args[0]

		fs := filesystems.FsLocalFile{Path: path}
		if fs.Exists() {
			if !force {
				return fmt.Errorf("%s already exists (use --force to overwrite)\n", path)
			}
		}

		payload := StoreConfig{
			Version: 0.1,
			Context: "default",
			Namespaces: map[string]map[string]Store{
				"default": {
					".env": {
						Type: ".env",
						Args: map[string]any{
							"input": ".env",
						},
					},
				},
			},
		}
		enc := &encoders.ObjToJsonEncoder{Newline: true}
		dumper := common.NewMarshaller(enc, payload)
		w, err := fs.OpenWriter()
		if err != nil {
			return err
		}
		defer w.Close()
		err = dumper.Dump(w)
		return err
	},
}
