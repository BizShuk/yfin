package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootHelpDescribesCurrentLocalDataSurface(t *testing.T) {
	help := RootCmd.Short + "\n" + RootCmd.Long
	require.NotContains(t, strings.ToLower(help), "bus")
	require.NotContains(t, strings.ToLower(help), "publisher")
	require.NotContains(t, strings.ToLower(help), "publishing")
	require.Contains(t, help, "本地匯出")
}
