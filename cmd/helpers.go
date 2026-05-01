package cmd

import (
	"context"
	"fmt"

	"github.com/Ab-code520/cloud-cli/core"
)

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", float64(bytes)/1024/1024/1024)
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/1024/1024)
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// findObjectByPath is a simplified search.
// It currently only searches the root directory.
// TODO: Implement full path traversal for production use.
func findObjectByPath(driver core.Storage, path string) (*core.Object, error) {
	list, err := driver.List(context.Background(), "")
	if err != nil {
		return nil, err
	}
	for _, obj := range list {
		// Match by Name or Path (ID)
		if obj.Name == path || obj.Path == path || obj.ID == path {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("object not found: %s", path)
}
