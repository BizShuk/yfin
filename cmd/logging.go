// logging.go — root CLI log-level precedence and default slog setup.
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const defaultLogLevel = "info"

// ResolveLogLevel applies CLI > APP_LOG_LEVEL > YAML > built-in precedence.
func ResolveLogLevel(command *cobra.Command, configured string) string {
	if command != nil {
		if flag := command.Flag("log-level"); flag != nil && flag.Changed {
			return flag.Value.String()
		}
	}
	if environmentLevel := strings.TrimSpace(os.Getenv("APP_LOG_LEVEL")); environmentLevel != "" {
		return environmentLevel
	}
	if configuredLevel := strings.TrimSpace(configured); configuredLevel != "" {
		return configuredLevel
	}
	return defaultLogLevel
}

func configureDefaultLogger(levelText string) error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(levelText)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", levelText, err)
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
	return nil
}
