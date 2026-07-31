package main

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSoakLoggerHonorsConfiguredLevel(t *testing.T) {
	logger, err := newSoakLogger("warn")

	require.NoError(t, err)
	require.False(t, logger.Enabled(context.Background(), slog.LevelInfo))
	require.True(t, logger.Enabled(context.Background(), slog.LevelWarn))
}

func TestNewSoakLoggerRejectsUnknownLevel(t *testing.T) {
	_, err := newSoakLogger("verbose")

	require.ErrorContains(t, err, "log level")
}
