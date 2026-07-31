// logging.go — stdlib structured logger construction for the soak binary.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func newSoakLogger(levelName string) (*slog.Logger, error) {
	var level slog.Level
	switch strings.ToLower(strings.TrimSpace(levelName)) {
	case "debug":
		level = slog.LevelDebug
	case "", "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("unsupported log level %q", levelName)
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler).With("service", "yfin-soak"), nil
}
