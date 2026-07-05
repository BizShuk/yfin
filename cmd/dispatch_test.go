// dispatch_test.go — single registry-coverage test asserting all 30 expected command names are present in `commandRegistry`. Capacity: 1 test function over a 30-name allow-list.
package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandRegistry_CoversAllCommands(t *testing.T) {
	want := []string{
		"info", "history", "actions", "income", "balance", "cashflow",
		"major-holders", "institutional-holders", "mutualfund-holders",
		"insider-transactions", "insider-purchases", "insider-roster",
		"recommendations", "recommendations-summary", "upgrades",
		"earnings-dates", "earnings-history", "eps-trend", "eps-revisions",
		"earnings-estimates", "revenue-estimates", "growth-estimates",
		"price-targets", "news", "calendar", "sec-filings", "sustainability",
		"isin", "options", "metadata",
	}
	for _, cmd := range want {
		_, ok := commandRegistry[cmd]
		require.Truef(t, ok, "command %q missing from registry", cmd)
	}
	require.Len(t, commandRegistry, len(want))
}
