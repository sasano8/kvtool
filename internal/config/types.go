package config

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