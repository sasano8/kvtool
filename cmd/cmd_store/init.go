package cmd_store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sasano8/kvtool/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	cmd := CmdInit
	cmd.Flags().BoolP("force", "f", false, "overwrite if file already exists")
	cmd.Flags().BoolP("global", "g", false, "create global config in ~/.config/kvtool/.kvtool.yml")
}

var CmdInit = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize kvtool configuration file",
	Long: `Initialize kvtool configuration file.

By default, creates .kvtool.yml in the current directory.
Use --global to create a global config in ~/.config/kvtool/.kvtool.yml

Examples:
  kvtool store init                    # Create .kvtool.yml in current directory
  kvtool store init --global           # Create global config
  kvtool store init myconfig.yml       # Create config at specified path
  kvtool store init --force            # Overwrite existing config`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		global, _ := cmd.Flags().GetBool("global")

		// Determine config path
		var configPath string
		if len(args) > 0 {
			configPath = args[0]
		} else if global {
			globalPath, err := config.GetGlobalConfigPath()
			if err != nil {
				return err
			}
			configPath = globalPath
		} else {
			configPath = ".kvtool.yml"
		}

		// Check if file already exists
		if _, err := os.Stat(configPath); err == nil {
			if !force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", configPath)
			}
		}

		// Create parent directory if needed
		dir := filepath.Dir(configPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}

		// Create default config
		config := config.KvtoolConfig{
			Version: 0.1,
			Namespaces: map[string]map[string]config.StoreInfo{
				"default": {
					".env": {
						Driver: "local",
						Args: map[string]interface{}{
							"transform": map[string]interface{}{
								"read":  "dotenv",
								"write": "dotenv",
							},
							"root": ".",
						},
						Mount: &config.MountInfo{
							File: stringPtr(""),
						},
					},
				},
			},
		}

		// Write config as YAML
		data, err := yaml.Marshal(&config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		absPath, _ := filepath.Abs(configPath)
		fmt.Printf("Created config file: %s\n", absPath)
		return nil
	},
}

func stringPtr(s string) *string {
	return &s
}
