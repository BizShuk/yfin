// dispatch_test.go — minimal sanity check that every entry in commandRegistry
// has a non-nil fetcher. (More thorough coverage would mock FetchContext; for
// now we just confirm the wiring is complete.)
package dispatch

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandRegistry_CoversAllCommands(t *testing.T) {
	require.NotEmpty(t, commandRegistry)
	for name, fn := range commandRegistry {
		require.NotNil(t, fn, "command %q has nil fetcher", name)
	}
}

func TestMigratedCommandsUseYahooEndpoints(t *testing.T) {
	source, err := os.ReadFile("dispatch.go")
	require.NoError(t, err)
	text := string(source)

	for _, call := range []string{
		"fc.Root.FetchIncomeStatement(ctx, s)",
		"fc.Root.FetchBalanceSheet(ctx, s)",
		"fc.Root.FetchCashFlowStatement(ctx, s)",
		"fc.Root.FetchNews(ctx, s)",
	} {
		require.Contains(t, text, call)
	}
	for _, legacyCall := range []string{
		"fc.Root.ScrapeFinancials(ctx, s, fc.RunID)",
		"fc.Root.ScrapeBalanceSheet(ctx, s, fc.RunID)",
		"fc.Root.ScrapeCashFlow(ctx, s, fc.RunID)",
		"fc.Root.ScrapeNews(ctx, s, fc.RunID)",
	} {
		require.NotContains(t, text, legacyCall)
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
