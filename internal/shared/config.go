package shared

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// KvtoolConfig represents the structure of .kvtool.yml
type KvtoolConfig struct {
	Version    float64                         `yaml:"version"`
	Namespaces map[string]map[string]StoreInfo `yaml:"namespaces"`
}

// StoreInfo represents a single store configuration
type StoreInfo struct {
	Driver string                 `yaml:"driver"`
	Args   map[string]interface{} `yaml:"args"`
	Mount  *MountInfo             `yaml:"mount,omitempty"`
}

// MountInfo represents mount configuration
type MountInfo struct {
	Dir  *string `yaml:"dir,omitempty"`
	File *string `yaml:"file,omitempty"`
}

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

// GetStore retrieves store configuration by namespace and store name
func (c *KvtoolConfig) GetStore(namespace, storeName string) (*StoreInfo, error) {
	ns, ok := c.Namespaces[namespace]
	if !ok {
		return nil, fmt.Errorf("namespace %q not found", namespace)
	}

	store, ok := ns[storeName]
	if !ok {
		return nil, fmt.Errorf("store %q not found in namespace %q", storeName, namespace)
	}

	return &store, nil
}

// StorePath represents a parsed store path
type StorePath struct {
	StoreName string
	FilePath  string
}

// ParseStorePath parses a path in the format "store_name/file_path"
// Returns the store name and file path
func ParseStorePath(path string) (*StorePath, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	parts := splitPath(path)
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid path format: expected store_name[/file_path], got %q", path)
	}

	result := &StorePath{
		StoreName: parts[0],
	}

	if len(parts) > 1 {
		result.FilePath = joinPath(parts[1:])
	}

	return result, nil
}

// splitPath splits a path by "/" and filters out empty parts
func splitPath(path string) []string {
	var parts []string
	current := ""

	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(path[i])
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// joinPath joins path parts with "/"
func joinPath(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "/"
		}
		result += part
	}
	return result
}
