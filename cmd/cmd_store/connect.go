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
	cmd := CmdConnect
	cmd.Flags().StringP("config", "c", ".kvtool.yml", "path to config file")
	cmd.Flags().BoolP("global", "g", false, "use global config (~/.config/kvtool/.kvtool.yml)")
	cmd.Flags().StringP("namespace", "n", "default", "namespace to use")
	cmd.Flags().DurationP("timeout", "t", 0, "overall timeout (0 = wait forever)")
	cmd.Flags().Duration("interval", 2*time.Second, "retry interval between connection attempts")
}

var CmdConnect = &cobra.Command{
	Use:   "connect <store_name|all>",
	Short: "Test connection to stores",
	Long: `Test connection to configured stores.

Retries until all specified stores are accessible or timeout is reached.
For local filesystems, this always succeeds immediately.
For remote stores (Vault, S3, NATS, Redis), this attempts to connect and verify access.

By default, retries forever (timeout=0). Use -t to set an upper limit.
Can be used as a dependency check to replace docker-compose depends_on.

Examples:
  kvtool store connect all                # Wait forever until all stores are ready
  kvtool store connect vault              # Wait for specific store
  kvtool store connect all -t 30s         # Timeout after 30 seconds
  kvtool store connect all -n production  # Test stores in production namespace`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		useGlobal, _ := cmd.Flags().GetBool("global")
		namespace, _ := cmd.Flags().GetString("namespace")
		timeout, _ := cmd.Flags().GetDuration("timeout")
		interval, _ := cmd.Flags().GetDuration("interval")

		// If --global flag is set, use global config
		if useGlobal {
			globalPath, err := config.GetGlobalConfigPath()
			if err != nil {
				return err
			}
			configPath = globalPath
		}

		// Load configuration
		cfg, err := config.LoadConfigAuto(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get namespace
		ns, exists := cfg.Namespaces[namespace]
		if !exists {
			return fmt.Errorf("namespace '%s' not found in config", namespace)
		}

		// Determine which stores to test
		storeName := args[0]
		var storesToTest []string
		if storeName == config.ReservedStoreAll {
			for name := range ns {
				storesToTest = append(storesToTest, name)
			}
		} else {
			if _, exists := ns[storeName]; !exists {
				return fmt.Errorf("store '%s' not found in namespace '%s'", storeName, namespace)
			}
			storesToTest = []string{storeName}
		}

		return connectStores(ns, storesToTest, namespace, timeout, interval)
	},
}

func connectStores(ns map[string]config.StoreInfo, storesToTest []string, namespace string, timeout, interval time.Duration) error {
	if timeout == 0 {
		fmt.Printf("Waiting for stores in namespace '%s' to be ready...\n\n", namespace)
	} else {
		fmt.Printf("Waiting for stores in namespace '%s' to be ready (timeout: %s)...\n\n", namespace, timeout)
	}

	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	for {
		allSuccess := true
		for _, storeName := range storesToTest {
			storeInfo := ns[storeName]

			attemptCtx, cancel := context.WithTimeout(context.Background(), interval)
			result := testConnection(attemptCtx, storeName, &storeInfo)
			cancel()

			if !result.Success {
				allSuccess = false
			}
		}

		if allSuccess {
			for _, storeName := range storesToTest {
				storeInfo := ns[storeName]
				fmt.Printf("✓ %s (%s)\n", storeName, storeInfo.Driver)
			}
			fmt.Printf("\nAll %d store(s) ready.\n", len(storesToTest))
			return nil
		}

		// timeout > 0 の場合のみデッドラインチェック
		if timeout > 0 && time.Now().After(deadline) {
			failCount := 0
			for _, storeName := range storesToTest {
				storeInfo := ns[storeName]
				attemptCtx, cancel := context.WithTimeout(context.Background(), interval)
				result := testConnection(attemptCtx, storeName, &storeInfo)
				cancel()
				if result.Success {
					fmt.Printf("✓ %s (%s): %s\n", storeName, storeInfo.Driver, result.Message)
				} else {
					fmt.Printf("✗ %s (%s): %s\n", storeName, storeInfo.Driver, result.Message)
					failCount++
				}
			}
			return fmt.Errorf("timeout: %d store(s) not ready after %s", failCount, timeout)
		}

		time.Sleep(interval)
	}
}

type ConnectionResult struct {
	Success bool
	Message string
}

func testConnection(ctx context.Context, name string, storeInfo *config.StoreInfo) ConnectionResult {
	// Use FilesystemFactory to create filesystem
	// This performs driver-specific validation (e.g., HeadBucket for S3)
	factory := filesystems.NewFilesystemFactory(ctx)

	fsStoreInfo := &filesystems.StoreInfo{
		Driver: storeInfo.Driver,
		Args:   storeInfo.Args,
	}

	_, err := factory.Create(fsStoreInfo)
	if err != nil {
		return ConnectionResult{
			Success: false,
			Message: fmt.Sprintf("Failed - %v", err),
		}
	}

	// Build success message based on driver type
	var details string
	switch storeInfo.Driver {
	case "local":
		details = "local filesystem"
		if root, ok := storeInfo.Args["root"].(string); ok && root != "" {
			details = fmt.Sprintf("local filesystem (root: %s)", root)
		}
	case "env":
		details = "environment variables"
	case "s3":
		bucket, _ := storeInfo.Args["bucket"].(string)
		endpoint, _ := storeInfo.Args["endpoint"].(string)
		if endpoint == "" {
			endpoint = "AWS S3"
		}
		details = fmt.Sprintf("bucket: %s, endpoint: %s", bucket, endpoint)
	case "vault":
		addr, _ := storeInfo.Args["addr"].(string)
		mount, _ := storeInfo.Args["mount"].(string)
		details = fmt.Sprintf("addr: %s, mount: %s", addr, mount)
	case "nats":
		url, _ := storeInfo.Args["url"].(string)
		bucket, _ := storeInfo.Args["bucket"].(string)
		details = fmt.Sprintf("url: %s, bucket: %s", url, bucket)
	case "redis":
		addr, _ := storeInfo.Args["addr"].(string)
		db, _ := storeInfo.Args["db"].(int)
		details = fmt.Sprintf("addr: %s, db: %d", addr, db)
	default:
		details = "connected"
	}

	return ConnectionResult{
		Success: true,
		Message: fmt.Sprintf("OK (%s)", details),
	}
}
