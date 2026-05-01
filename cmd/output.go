package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// outputFormat controls the global output format (table, json).
var outputFormat string

// outputJSON encodes v as JSON to stdout.
func outputJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// outputTableOrJSON calls tableFn if outputFormat is "table" (default),
// otherwise encodes v as JSON.
func outputTableOrJSON(v interface{}, tableFn func()) error {
	switch strings.ToLower(outputFormat) {
	case "json":
		return outputJSON(v)
	default:
		tableFn()
		return nil
	}
}

func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
