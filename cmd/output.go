package cmd

import (
	"encoding/json"
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
