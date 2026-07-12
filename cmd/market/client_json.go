// client_json.go — local-export sink shared between `pull` and `quote`.
// Writes any JSON-serialisable value to filepath with `  ` indented output.
// Both subcommands' `handleLocalExport` / `handleQuoteLocalExport` route
// their `--out json` writes through this helper.
package market

import (
	"encoding/json"
	"os"
)

// writeJSONFile writes data to a JSON file
func writeJSONFile(filepath string, data interface{}) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
