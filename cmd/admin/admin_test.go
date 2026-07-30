package admin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigHelpUsesCurrentCommandAndFlags(t *testing.T) {
	cmd := newConfigCmd()
	help := cmd.Use + "\n" + cmd.Long

	require.Equal(t, "config", cmd.Use)
	require.NotContains(t, strings.ToLower(help), "ampy-config")
	require.NotContains(t, help, "--file")
	require.Contains(t, help, "yfin --config ./config/effective.yaml config --print-effective")
}
