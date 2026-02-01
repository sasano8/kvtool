package cmd_store

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sasano8/kvtool/filesystems"
	"github.com/sasano8/kvtool/internal/shared"
	"github.com/sasano8/kvtool/pkg/decoders"
	"github.com/sasano8/kvtool/pkg/encoders"
	"github.com/spf13/cobra"
)

func init() {
	cmd := CmdGet
	cmd.Flags().StringP("config", "c", ".kvtool.yml", "path to config file")
	cmd.Flags().BoolP("global", "g", false, "use global config (~/.config/kvtool/.kvtool.yml)")
	cmd.Flags().StringP("output", "o", "json", "output format (json, yaml, raw)")
	cmd.Flags().StringP("namespace", "n", "default", "namespace to use")
}

var CmdGet = &cobra.Command{
	Use:   "get <store_name>/<file_path>",
	Short: "Get a file from the store",
	Long: `Get a file from the store.

Examples:
  kvtool store get .env/APP_NAME
  kvtool store get vault/app/prod/db_password
  kvtool store get .env/APP_NAME -o raw
  kvtool store get .env/APP_NAME --namespace production
  kvtool store get .env/APP_NAME --global`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		useGlobal, _ := cmd.Flags().GetBool("global")
		outputFormat, _ := cmd.Flags().GetString("output")
		namespace, _ := cmd.Flags().GetString("namespace")
		storePath := args[0]

		// If --global flag is set, use global config
		if useGlobal {
			globalPath, err := shared.GetGlobalConfigPath()
			if err != nil {
				return err
			}
			configPath = globalPath
		}

		// Parse the store path
		parsed, err := shared.ParseStorePath(storePath)
		if err != nil {
			return err
		}

		// Load config with automatic path resolution
		config, err := shared.LoadConfigAuto(configPath)
		if err != nil {
			return err
		}

		// Get store info
		storeInfo, err := config.GetStore(namespace, parsed.StoreName)
		if err != nil {
			return err
		}

		// Get file content based on driver
		var content interface{}
		switch storeInfo.Driver {
		case "local":
			content, err = getFromLocalFs(storeInfo, parsed.FilePath)
		case "vault":
			content, err = getFromVaultFs(storeInfo, parsed.FilePath)
		default:
			return fmt.Errorf("unsupported driver: %s", storeInfo.Driver)
		}

		if err != nil {
			return err
		}

		// Output based on format
		return outputContent(content, outputFormat, os.Stdout)
	},
}

func getFromLocalFs(storeInfo *shared.StoreInfo, filePath string) (interface{}, error) {
	args := storeInfo.Args
	root, _ := args["root"].(string)
	if root == "" {
		root = "."
	}

	ctx := context.Background()
	fs, err := filesystems.GetLocalFs(ctx, &filesystems.LocalFsConfig{
		Root:    root,
		Timeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// Check for transform
	transform := getTransform(args, "read")

	if transform == "dotenv" {
		reader, err := fs.OpenReader(filePath)
		if err != nil {
			return nil, err
		}
		defer reader.Close()

		return decoders.DotenvToJson(reader)
	}

	// Default: load as JSON
	return fs.LoadAsJson(filePath)
}

func getFromVaultFs(storeInfo *shared.StoreInfo, filePath string) (interface{}, error) {
	args := storeInfo.Args

	// Parse vault connection args
	connArgs, ok := args["conn"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("vault store missing 'conn' configuration")
	}

	addr, _ := connArgs["addr"].(string)
	token, _ := connArgs["token"].(string)
	mount, _ := connArgs["mount"].(string)

	if addr == "" || token == "" || mount == "" {
		return nil, fmt.Errorf("vault store missing required configuration (addr, token, mount)")
	}

	root, _ := args["root"].(string)
	if root != "" {
		filePath = root + "/" + filePath
	}

	config := filesystems.VaultConfig{
		Addr:      addr,
		Token:     token,
		Namespace: "admin",
		Mount:     mount,
		KvVer:     2,
		Version:   0,
		Timeout:   10 * time.Second,
	}

	ctx := context.Background()
	vaultFs, err := filesystems.GetVaultFs(ctx, &config)
	if err != nil {
		return nil, err
	}

	file, err := vaultFs.GetFile(filePath)
	if err != nil {
		return nil, err
	}

	return file.LoadAsJson()
}

func getTransform(args map[string]interface{}, direction string) string {
	transform, ok := args["transform"].(map[string]interface{})
	if !ok {
		return ""
	}

	val, _ := transform[direction].(string)
	return val
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
