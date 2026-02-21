package cmd_store

import (
	"fmt"
	"sort"

	"github.com/sasano8/kvtool/pkg/config"
	"github.com/spf13/cobra"
)

func init() {
	cmd := CmdTree
	cmd.Flags().StringP("config", "c", ".kvtool.yml", "path to config file")
	cmd.Flags().BoolP("global", "g", false, "use global config (~/.config/kvtool/.kvtool.yml)")
	cmd.Flags().StringP("namespace", "n", "default", "namespace to use")
}

var CmdTree = &cobra.Command{
	Use:   "tree <store_name|all>",
	Short: "Display store configuration as a tree",
	Long: `Display the logical hierarchy of store configurations.

Examples:
  kvtool store tree all                # Show all stores in default namespace
  kvtool store tree vault              # Show specific store
  kvtool store tree all -n production  # Show stores in production namespace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		useGlobal, _ := cmd.Flags().GetBool("global")
		namespace, _ := cmd.Flags().GetString("namespace")

		if useGlobal {
			globalPath, err := config.GetGlobalConfigPath()
			if err != nil {
				return err
			}
			configPath = globalPath
		}

		cfg, err := config.LoadConfigAuto(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		ns, exists := cfg.Namespaces[namespace]
		if !exists {
			return fmt.Errorf("namespace '%s' not found in config", namespace)
		}

		storeName := args[0]
		var storeNames []string
		if storeName == config.ReservedStoreAll {
			for name := range ns {
				storeNames = append(storeNames, name)
			}
			sort.Strings(storeNames)
		} else {
			if _, exists := ns[storeName]; !exists {
				return fmt.Errorf("store '%s' not found in namespace '%s'", storeName, namespace)
			}
			storeNames = []string{storeName}
		}

		printTree(namespace, ns, storeNames)
		return nil
	},
}

func printTree(namespace string, ns map[string]config.StoreInfo, storeNames []string) {
	fmt.Printf("Namespace: %s\n", namespace)

	for i, name := range storeNames {
		store := ns[name]
		isLast := i == len(storeNames)-1

		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		fmt.Printf("%s %s (driver: %s)\n", prefix, name, store.Driver)
	}
}
