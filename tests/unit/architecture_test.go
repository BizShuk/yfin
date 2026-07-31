package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const modulePath = "github.com/bizshuk/yfin"

type listedPackage struct {
	ImportPath string
	Imports    []string
}

func TestInternalPackagesFollowDependencyContract(t *testing.T) {
	root := repositoryRoot(t)
	command := exec.Command("go", "list", "-json", "./...")
	command.Dir = root
	output, err := command.CombinedOutput()
	require.NoErrorf(t, err, "go list failed:\n%s", output)

	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var pkg listedPackage
		err = decoder.Decode(&pkg)
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)

		for _, imported := range pkg.Imports {
			if !strings.HasPrefix(imported, modulePath+"/") {
				continue
			}
			require.Truef(
				t,
				internalImportAllowed(pkg.ImportPath, imported),
				"%s must not import %s",
				pkg.ImportPath,
				imported,
			)
		}
	}
}

func internalImportAllowed(pkg, imported string) bool {
	switch {
	case pkg == modulePath:
		return strings.HasPrefix(imported, modulePath+"/cmd")
	case pkg == modulePath+"/model":
		return false
	case pkg == modulePath+"/config":
		return imported == modulePath+"/utils/httpx"
	case pkg == modulePath+"/utils/httpx":
		return imported == modulePath+"/utils/obsv"
	case strings.HasPrefix(pkg, modulePath+"/utils/"):
		return false
	case strings.HasPrefix(pkg, modulePath+"/svc/"):
		return imported == modulePath+"/model" || imported == modulePath+"/utils/httpx"
	case pkg == modulePath+"/facade":
		return imported == modulePath+"/config" ||
			imported == modulePath+"/model" ||
			imported == modulePath+"/utils/cache" ||
			imported == modulePath+"/utils/httpx" ||
			strings.HasPrefix(imported, modulePath+"/svc/")
	case pkg == modulePath+"/cmd/format":
		return imported == modulePath+"/model"
	case pkg == modulePath+"/cmd/soak":
		return imported == modulePath+"/config" ||
			imported == modulePath+"/facade" ||
			imported == modulePath+"/model" ||
			imported == modulePath+"/utils/httpx"
	case strings.HasPrefix(pkg, modulePath+"/cmd"):
		return imported == modulePath+"/cmd" ||
			imported == modulePath+"/cmd/format" ||
			imported == modulePath+"/config" ||
			imported == modulePath+"/facade" ||
			imported == modulePath+"/model" ||
			imported == modulePath+"/utils/cache" ||
			imported == modulePath+"/utils/httpx" ||
			imported == modulePath+"/utils/obsv"
	default:
		return true
	}
}
