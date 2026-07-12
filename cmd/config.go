// config.go — `config` cobra subcommand (`--print-effective` dumps the
// resolved ampy-config to stdout) plus the `printEffectiveConfig` /
// `flattenConfigMap` helpers that turn a nested map into dot-notation
// key=value lines. Capacity: 1 `ConfigConfig` + 1 var + 1 `configCmd` +
// 1 `init()` (2 flags) + `runConfig` / `printEffectiveConfig` /
// `flattenConfigMap`.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bizshuk/yfin/config"
	"github.com/spf13/cobra"
)

// ConfigConfig holds configuration for the config command
type ConfigConfig struct {
	PrintEffective bool
	JSON           bool
}

var configConfig ConfigConfig

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long: `Configuration management for yfinance-go.
Loads and validates configuration from ampy-config files.

Examples:
  yfin config --file ./config/example.dev.yaml --print-effective
  yfin config --print-effective --json`,
	RunE: runConfig,
}

func init() {
	configCmd.Flags().BoolVar(&configConfig.PrintEffective, "print-effective", false, "Print effective configuration")
	configCmd.Flags().BoolVar(&configConfig.JSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(configCmd)
}

// runConfig executes the config command
func runConfig(cmd *cobra.Command, args []string) error {
	if !configConfig.PrintEffective {
		return fmt.Errorf("--print-effective flag is required")
	}

	// Determine effective config path
	effectivePath := globalConfig.ConfigFile
	if effectivePath == "" {
		// Default to a standard effective config path
		effectivePath = "config/effective.yaml"
	}

	// Load configuration using ampy-config
	loader := config.NewLoader(effectivePath)
	_, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Get effective configuration
	effectiveConfig, err := loader.GetEffectiveConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to get effective configuration: %v\n", err)
		os.Exit(ExitConfigError)
	}

	// Print configuration
	if configConfig.JSON {
		// Print as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(effectiveConfig); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to encode configuration as JSON: %v\n", err)
			os.Exit(ExitConfigError)
		}
	} else {
		// Print as key=value pairs
		printEffectiveConfig(effectiveConfig)
	}

	return nil
}

// printEffectiveConfig prints the effective configuration in key=value format
func printEffectiveConfig(configMap map[string]interface{}) {
	fmt.Println("EFFECTIVE CONFIG (redacted)")

	// Flatten the configuration map
	flattened := flattenConfigMap(configMap, "")

	// Sort keys for consistent output
	var keys []string
	for key := range flattened {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Print sorted key-value pairs
	for _, key := range keys {
		value := flattened[key]
		fmt.Printf("%s=%v\n", key, value)
	}
}

// flattenConfigMap flattens a nested configuration map into dot-notation keys
func flattenConfigMap(configMap map[string]interface{}, prefix string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range configMap {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Recursively flatten nested maps
			nested := flattenConfigMap(nestedMap, fullKey)
			for k, v := range nested {
				result[k] = v
			}
		} else if slice, ok := value.([]interface{}); ok {
			// Handle slices (like allowed_intervals)
			var strSlice []string
			for _, item := range slice {
				if str, ok := item.(string); ok {
					strSlice = append(strSlice, str)
				}
			}
			if len(strSlice) > 0 {
				result[fullKey] = fmt.Sprintf("[%s]", strings.Join(strSlice, ","))
			}
		} else {
			result[fullKey] = value
		}
	}

	return result
}
