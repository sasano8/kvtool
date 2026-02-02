package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads .kvtool.yml from the specified path
func LoadConfig(path string) (*KvtoolConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config KvtoolConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// ResolveConfigPath resolves the config file path with the following priority:
// 1. Explicitly specified path (if not empty and not default)
// 2. Local .kvtool.yml in current directory (no fallback to global)
// Returns the resolved path and nil error if found, or error if not found
func ResolveConfigPath(explicitPath string) (string, error) {
	// 1. If explicitly specified, use that
	if explicitPath != "" && explicitPath != ".kvtool.yml" {
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath, nil
		}
		return "", fmt.Errorf("config file not found at specified path: %s", explicitPath)
	}

	// 2. Check local .kvtool.yml (no fallback to global)
	localPath := ".kvtool.yml"
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	return "", fmt.Errorf("config file not found: %s (run 'kvtool store init' to create one)", localPath)
}

// GetGlobalConfigPath returns the path to the global config file
// Uses ~/.config/kvtool/.kvtool.yml
func GetGlobalConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "kvtool", ".kvtool.yml"), nil
}

// LoadConfigAuto loads config with automatic path resolution
func LoadConfigAuto(explicitPath string) (*KvtoolConfig, error) {
	path, err := ResolveConfigPath(explicitPath)
	if err != nil {
		return nil, err
	}

	return LoadConfig(path)
}
