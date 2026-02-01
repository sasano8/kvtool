package config

import "fmt"

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