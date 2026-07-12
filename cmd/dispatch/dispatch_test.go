// dispatch_test.go — minimal sanity check that every entry in commandRegistry
// has a non-nil fetcher. (More thorough coverage would mock FetchContext; for
// now we just confirm the wiring is complete.)
package dispatch

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandRegistry_CoversAllCommands(t *testing.T) {
	require.NotEmpty(t, commandRegistry)
	for name, fn := range commandRegistry {
		require.NotNil(t, fn, "command %q has nil fetcher", name)
	}
}
