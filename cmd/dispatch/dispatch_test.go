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

func TestCommandRegistryMatchesPythonManifest(t *testing.T) {
	want := []string{
		"info",
		"history",
		"actions",
		"income",
		"balance",
		"cashflow",
		"major-holders",
		"institutional-holders",
		"mutualfund-holders",
		"insider-transactions",
		"insider-purchases",
		"insider-roster",
		"recommendations",
		"recommendations-summary",
		"upgrades",
		"earnings-dates",
		"earnings-history",
		"eps-trend",
		"eps-revisions",
		"earnings-estimates",
		"revenue-estimates",
		"growth-estimates",
		"price-targets",
		"news",
		"calendar",
		"sec-filings",
		"sustainability",
		"isin",
		"options",
		"metadata",
	}

	require.Len(t, commandRegistry, len(want))
	require.Equal(t, want, commandOrder)
	for _, command := range commandOrder {
		require.Contains(t, commandRegistry, command)
	}
}
