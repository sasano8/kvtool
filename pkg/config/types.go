package config

import "fmt"

// ReservedStoreAll は全ストアを指定する予約語
const ReservedStoreAll = "all"

// ReservedStoreNames はストア名として使用できない予約語一覧
var ReservedStoreNames = []string{ReservedStoreAll}

// KvtoolConfig represents the structure of .kvtool.yml
type KvtoolConfig struct {
	Version    float64                         `yaml:"version"`
	Namespaces map[string]map[string]StoreInfo `yaml:"namespaces"`
}

// Validate はコンフィグのバリデーションを行う
func (c *KvtoolConfig) Validate() error {
	for nsName, ns := range c.Namespaces {
		for storeName := range ns {
			for _, reserved := range ReservedStoreNames {
				if storeName == reserved {
					return fmt.Errorf("store name %q is a reserved word in namespace %q", storeName, nsName)
				}
			}
		}
	}
	return nil
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
