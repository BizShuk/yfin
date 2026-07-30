package cmd

import (
	"context"
	"log/slog"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestResolveLogLevelPrecedence(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	command.Flags().String("log-level", "info", "")

	t.Setenv("APP_LOG_LEVEL", "warn")
	require.Equal(t, "warn", ResolveLogLevel(command, "error"))

	require.NoError(t, command.Flags().Set("log-level", "debug"))
	require.Equal(t, "debug", ResolveLogLevel(command, "error"))
}

func TestResolveLogLevelFallsBackToYAMLThenInfo(t *testing.T) {
	t.Setenv("APP_LOG_LEVEL", "")

	command := &cobra.Command{Use: "test"}
	command.Flags().String("log-level", "info", "")

	require.Equal(t, "error", ResolveLogLevel(command, "error"))
	require.Equal(t, "info", ResolveLogLevel(command, ""))
}

func TestResolveLogLevelReadsInheritedPersistentFlag(t *testing.T) {
	t.Setenv("APP_LOG_LEVEL", "warn")

	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("log-level", "info", "")
	child := &cobra.Command{
		Use: "child",
		RunE: func(command *cobra.Command, _ []string) error {
			require.Equal(t, "debug", ResolveLogLevel(command, "error"))
			return nil
		},
	}
	root.AddCommand(child)
	root.SetArgs([]string{"--log-level", "debug", "child"})

	require.NoError(t, root.Execute())
}

func TestConfigureDefaultLoggerAppliesAndValidatesLevel(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(previous)
	})

	require.NoError(t, configureDefaultLogger("debug"))
	require.True(t, slog.Default().Enabled(context.Background(), slog.LevelDebug))

	require.ErrorContains(t, configureDefaultLogger("verbose"), "invalid log level")
}
