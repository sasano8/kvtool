package cmd_store

import (
	"context"
	"fmt"
	"time"

	"github.com/sasano8/kvtool/pkg/config"
	"github.com/sasano8/kvtool/pkg/filesystems"
	"github.com/spf13/cobra"
)

func init() {
	cmd := CmdSync
	cmd.Flags().StringP("config", "c", ".kvtool.yml", "path to config file")
	cmd.Flags().BoolP("global", "g", false, "use global config (~/.config/kvtool/.kvtool.yml)")
	cmd.Flags().StringP("namespace", "n", "default", "namespace to use")
	cmd.Flags().DurationP("timeout", "t", 60*time.Second, "overall timeout for sync operations")
}

var CmdSync = &cobra.Command{
	Use:   "sync <store_name|all>",
	Short: "Sync remote stores to local cache",
	Long: `Sync remote stores (e.g., git) to local cache.

For stores that implement Syncable (e.g., git), this pulls the latest data.
Non-syncable stores are skipped silently.

Examples:
  kvtool store sync all                # Sync all syncable stores
  kvtool store sync gitconfig          # Sync specific store
  kvtool store sync all -n production  # Sync stores in production namespace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		useGlobal, _ := cmd.Flags().GetBool("global")
		namespace, _ := cmd.Flags().GetString("namespace")
		timeout, _ := cmd.Flags().GetDuration("timeout")

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
		var storesToSync []string
		if storeName == config.ReservedStoreAll {
			for name := range ns {
				storesToSync = append(storesToSync, name)
			}
		} else {
			if _, exists := ns[storeName]; !exists {
				return fmt.Errorf("store '%s' not found in namespace '%s'", storeName, namespace)
			}
			storesToSync = []string{storeName}
		}

		return syncStores(ns, storesToSync, namespace, timeout)
	},
}

func syncStores(ns map[string]config.StoreInfo, storesToSync []string, namespace string, timeout time.Duration) error {
	fmt.Printf("Syncing stores in namespace '%s'...\n\n", namespace)

	syncCount := 0
	skipCount := 0
	failCount := 0

	for _, storeName := range storesToSync {
		storeInfo := ns[storeName]

		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		factory := filesystems.NewFilesystemFactory(ctx)
		fsStoreInfo := &filesystems.StoreInfo{
			Driver: storeInfo.Driver,
			Args:   storeInfo.Args,
		}
		if storeInfo.Context != nil {
			fsStoreInfo.Context = &filesystems.ContextInfo{
				Timeout: storeInfo.Context.Timeout,
			}
		}

		fs, err := factory.Create(fsStoreInfo)
		if err != nil {
			cancel()
			fmt.Printf("✗ %s (%s): failed to create filesystem - %v\n", storeName, storeInfo.Driver, err)
			failCount++
			continue
		}

		syncable, ok := fs.(filesystems.Syncable)
		if !ok {
			cancel()
			fmt.Printf("- %s (%s): skipped (not syncable)\n", storeName, storeInfo.Driver)
			skipCount++
			continue
		}

		if err := syncable.Sync(); err != nil {
			cancel()
			fmt.Printf("✗ %s (%s): sync failed - %v\n", storeName, storeInfo.Driver, err)
			failCount++
			continue
		}

		cancel()
		fmt.Printf("✓ %s (%s): synced\n", storeName, storeInfo.Driver)
		syncCount++
	}

	fmt.Printf("\nSync complete: %d synced, %d skipped, %d failed\n", syncCount, skipCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("%d store(s) failed to sync", failCount)
	}

	return nil
}
