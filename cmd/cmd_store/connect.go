package cmd_store

import (
	"context"
	"fmt"
	"time"

	"github.com/sasano8/kvtool/pkg/filesystems"
	"github.com/sasano8/kvtool/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	cmd := CmdConnect
	cmd.Flags().StringP("config", "c", ".kvtool.yml", "path to config file")
	cmd.Flags().BoolP("global", "g", false, "use global config (~/.config/kvtool/.kvtool.yml)")
	cmd.Flags().StringP("namespace", "n", "default", "namespace to use")
	cmd.Flags().DurationP("timeout", "t", 10*time.Second, "connection timeout")
}

var CmdConnect = &cobra.Command{
	Use:   "connect [store_name]",
	Short: "Test connection to stores",
	Long: `Test connection to configured stores.

Verifies that stores are accessible and properly configured.
For local filesystems, this always succeeds.
For remote stores (Vault, S3), this attempts to connect and verify access.

Examples:
  kvtool store connect                    # Test all stores
  kvtool store connect vault              # Test specific store
  kvtool store connect -n production      # Test stores in production namespace`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, _ := cmd.Flags().GetString("config")
		useGlobal, _ := cmd.Flags().GetBool("global")
		namespace, _ := cmd.Flags().GetString("namespace")
		timeout, _ := cmd.Flags().GetDuration("timeout")

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

		// Filter stores if specific store is requested
		var storesToTest []string
		if len(args) > 0 {
			storeName := args[0]
			if _, exists := ns[storeName]; !exists {
				return fmt.Errorf("store '%s' not found in namespace '%s'", storeName, namespace)
			}
			storesToTest = []string{storeName}
		} else {
			// Test all stores
			for name := range ns {
				storesToTest = append(storesToTest, name)
			}
		}

		fmt.Printf("Testing connection to stores in namespace '%s':\n\n", namespace)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		successCount := 0
		failCount := 0

		for _, storeName := range storesToTest {
			storeInfo := ns[storeName]
			result := testConnection(ctx, storeName, &storeInfo)

			if result.Success {
				fmt.Printf("✓ %s (%s): %s\n", storeName, storeInfo.Driver, result.Message)
				successCount++
			} else {
				fmt.Printf("✗ %s (%s): %s\n", storeName, storeInfo.Driver, result.Message)
				failCount++
			}
		}

		fmt.Printf("\nSummary: %d succeeded, %d failed\n", successCount, failCount)

		if failCount > 0 {
			return fmt.Errorf("connection test failed for %d store(s)", failCount)
		}

		return nil
	},
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
	default:
		details = "connected"
	}

	return ConnectionResult{
		Success: true,
		Message: fmt.Sprintf("OK (%s)", details),
	}
}
