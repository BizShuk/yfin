package unit

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCLISampleScriptsUseCurrentSurfaces(t *testing.T) {
	root := repositoryRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, "cmd", "samples", "*.sh"))
	require.NoError(t, err)
	require.NotEmpty(t, paths)

	removedTokens := []string{
		"--preview-proto",
		"--publish",
		"--save-html",
		"--save-parsed",
		"--session-status",
		"--list-endpoints",
	}
	yfinSoakPattern := regexp.MustCompile(`(?m)(?:\byfin|\$YFIN_CMD)\s+soak\b`)

	for _, path := range paths {
		content, readErr := os.ReadFile(path)
		require.NoError(t, readErr)
		for _, token := range removedTokens {
			require.NotContainsf(t, string(content), token, "%s contains removed surface %q", path, token)
		}
		require.NotRegexpf(t, yfinSoakPattern, string(content), "%s treats standalone soak as a yfin subcommand", path)

		syntaxCheck := exec.Command("bash", "-n", path)
		output, syntaxErr := syntaxCheck.CombinedOutput()
		require.NoErrorf(t, syntaxErr, "%s is not valid shell: %s", path, output)
	}

	soakPath := filepath.Join(root, "cmd", "samples", "soak_smoke.sh")
	soakContent, err := os.ReadFile(soakPath)
	require.NoError(t, err)
	require.Contains(t, string(soakContent), "go run ./cmd/soak")
	require.NotContains(t, string(soakContent), "--preview")
}
