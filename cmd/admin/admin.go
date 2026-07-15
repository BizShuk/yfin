// admin.go — `config` + `version` cobra subcommands grouped under one
// sub-package because both are admin/maintenance commands (no network I/O).
// Capacity: 1 `Register(rootCmd)` exporting both commands.
package admin

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bizshuk/yfin/cmd"
	"github.com/bizshuk/yfin/config"
	"github.com/spf13/cobra"
)

// Register attaches the `config` and `version` subcommands onto rootCmd.
func Register(rootCmd *cobra.Command) {
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newVersionCmd())
}

// configCmd builders + RunE -----------------------------------------------

// ConfigConfig holds configuration for the config command
type configConfig struct {
	PrintEffective bool
	JSON           bool
}

// newConfigCmd returns the `config` cobra command. It dumps the resolved
// ampy-config (flattened to dot-notation key=value, or JSON) to stdout.
func newConfigCmd() *cobra.Command {
	cfg := &configConfig{}
	cmd := &cobra.Command{
		Use:   "config",
		Short: "組態管理 (Configuration management)",
		Long: `yfin CLI 組態管理：載入並驗證 ampy-config 檔。
(Configuration management for yfinance-go.
Loads and validates configuration from ampy-config files.)

範例 (Examples):
  yfin config --file ./config/effective.yaml --print-effective
  yfin config --print-effective --json`,
		RunE: func(c *cobra.Command, args []string) error { return runConfig(cfg) },
	}
	cmd.Flags().BoolVar(&cfg.PrintEffective, "print-effective", false, "Print effective configuration")
	cmd.Flags().BoolVar(&cfg.JSON, "json", false, "Output in JSON format")
	return cmd
}

// runConfig executes the config command
func runConfig(cfg *configConfig) error {
	if !cfg.PrintEffective {
		return fmt.Errorf("--print-effective flag is required")
	}

	// Determine effective config path
	effectivePath := cmd.Global.ConfigFile
	if effectivePath == "" {
		// Default to a standard effective config path
		effectivePath = "config/effective.yaml"
	}

	// Load configuration using ampy-config
	loader := config.NewLoader(effectivePath)
	_, err := loader.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to load configuration: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// Get effective configuration
	effectiveConfig, err := loader.GetEffectiveConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Failed to get effective configuration: %v\n", err)
		os.Exit(cmd.ExitConfigError)
	}

	// Print configuration
	if cfg.JSON {
		// Print as JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(effectiveConfig); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to encode configuration as JSON: %v\n", err)
			os.Exit(cmd.ExitConfigError)
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

// versionCmd builders + RunE ----------------------------------------------

// newVersionCmd returns the `version` cobra command. Reads cmd.Version /
// cmd.Commit / cmd.Date injected via `-ldflags`.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "列印版本資訊 (Print version information)",
		Long:  `列印 CLI 版本與 build 細節 (Print version information including build details).`,
		RunE: func(c *cobra.Command, args []string) error {
			fmt.Printf("yfin version %s\n", cmd.Version)
			fmt.Printf("commit: %s\n", cmd.Commit)
			fmt.Printf("build date: %s\n", cmd.Date)
			return nil
		},
	}
}
