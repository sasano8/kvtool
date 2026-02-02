package cmd_store

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sasano8/kvtool/pkg/config"
	"github.com/sasano8/kvtool/pkg/encoders"
	"github.com/sasano8/kvtool/pkg/service"
	"github.com/spf13/cobra"
)

func init() {
	cmd := CmdLoad
	cmd.Flags().StringP("config", "c", ".kvtool.yml", "path to config file")
	cmd.Flags().BoolP("global", "g", false, "use global config (~/.config/kvtool/.kvtool.yml)")
	cmd.Flags().StringP("output", "o", "json", "output format (json, yaml, raw)")
	cmd.Flags().StringP("namespace", "n", "default", "namespace to use")
}

var CmdLoad = &cobra.Command{
	Use:   "load <store_name>/<file_path>",
	Short: "Load a file from the store",
	Long: `Load a file from the store and output its content.

Examples:
  kvtool store load .env/APP_NAME
  kvtool store load vault/app/prod/db_password
  kvtool store load .env/APP_NAME -o raw
  kvtool store load .env/APP_NAME --namespace production
  kvtool store load .env/APP_NAME --global`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		useGlobal, _ := cmd.Flags().GetBool("global")
		outputFormat, _ := cmd.Flags().GetString("output")
		namespace, _ := cmd.Flags().GetString("namespace")
		storePath := args[0]

		// If --global flag is set, use global config
		if useGlobal {
			globalPath, err := config.GetGlobalConfigPath()
			if err != nil {
				return err
			}
			configPath = globalPath
		}

		// Use service to get content
		ctx := context.Background()
		storeService := service.NewStoreService()

		content, err := storeService.Get(ctx, service.GetOptions{
			ConfigPath: configPath,
			Namespace:  namespace,
			StorePath:  storePath,
		})
		if err != nil {
			return err
		}

		// Output based on format
		return outputContent(content, outputFormat, os.Stdout)
	},
}

func outputContent(content interface{}, format string, w io.Writer) error {
	switch format {
	case "json":
		encoder := encoders.ObjToJsonEncoder{Newline: true}
		bytes, err := encoder.Marshal(content)
		if err != nil {
			return err
		}
		_, err = w.Write(bytes)
		return err

	case "raw":
		// For raw output, try to convert to string
		switch v := content.(type) {
		case string:
			_, err := fmt.Fprint(w, v)
			return err
		case map[string]interface{}:
			// If it's a map, output as key=value pairs
			for key, val := range v {
				_, err := fmt.Fprintf(w, "%s=%v\n", key, val)
				if err != nil {
					return err
				}
			}
			return nil
		default:
			return fmt.Errorf("unsupported type for raw output: %T", content)
		}

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
