package unit

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunScriptInitializesYFinMetadataDirectories(t *testing.T) {
	configHome := t.TempDir()
	command := exec.Command(filepath.Join(repositoryRoot(t), "run.sh"))
	command.Env = append(os.Environ(), "XDG_CONFIG_HOME="+configHome)

	output, err := command.CombinedOutput()

	require.NoErrorf(t, err, "run.sh failed: %s", output)
	for _, name := range []string{"data", "logs"} {
		info, statErr := os.Stat(filepath.Join(configHome, "yfin", name))
		require.NoError(t, statErr)
		require.True(t, info.IsDir())
	}
}
