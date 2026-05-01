package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Ab-code520/cloud-cli/core"
)

// findObjectByPath resolves a path like "/folder/subfolder/file.txt" to an Object.
// It recursively traverses the directory structure from root ("0").
func findObjectByPath(driver core.Storage, path string) (*core.Object, error) {
	if path == "" || path == "/" || path == "0" {
		return &core.Object{ID: "0", Name: "/", IsDir: true}, nil
	}

	// Remove leading slash
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")
	
	// Find the file/dir
	return resolvePathParts(driver, context.Background(), "0", parts)
}

func resolvePathParts(driver core.Storage, ctx context.Context, currentID string, parts []string) (*core.Object, error) {
	if len(parts) == 0 {
		return driver.Info(ctx, currentID)
	}

	part := parts[0]
	
	// Optimization: If currentID is "0" and part is just a name, we can try Info if supported? 
	// No, Info usually takes ID. We must List.
	
	objs, err := driver.List(ctx, currentID)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", currentID, err)
	}

	for _, obj := range objs {
		if obj.Name == part {
			if len(parts) == 1 {
				// Found it
				return obj, nil
			}
			// It's a directory (intermediate part), recurse
			if obj.IsDir {
				return resolvePathParts(driver, ctx, obj.ID, parts[1:])
			} else {
				return nil, fmt.Errorf("%s is a file, not a directory", part)
			}
		}
	}

	return nil, fmt.Errorf("path '%s' not found", strings.Join(parts, "/"))
}

// findObjectByPathOrID tries to parse as ID first, then path.
func findObjectByPathOrID(driver core.Storage, path string) (*core.Object, error) {
	// If it looks like an ID (alphanumeric, length ~32), try Info first
	if isValidDirID(path) {
		obj, err := driver.Info(context.Background(), path)
		if err == nil {
			return obj, nil
		}
		// Fallback to path resolution
	}
	return findObjectByPath(driver, path)
}
