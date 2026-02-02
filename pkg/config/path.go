package config

import "fmt"

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
