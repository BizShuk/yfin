package facade

import (
	"context"
	"testing"
)

func TestFoundationCommandsMatchSupportedSurface(t *testing.T) {
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

	got := FoundationCommands()
	if len(got) != len(want) {
		t.Fatalf("FoundationCommands() len = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("FoundationCommands()[%d] = %q, want %q", index, got[index], want[index])
		}
	}

	got[0] = "mutated"
	if FoundationCommands()[0] != "info" {
		t.Fatal("FoundationCommands returned mutable package state")
	}
}

func TestFetchFoundationRejectsUnknownCommandBeforeProviderAccess(t *testing.T) {
	_, err := NewClient().FetchFoundation(
		context.Background(),
		"unknown",
		"AAPL",
		"test-run",
	)
	if err == nil {
		t.Fatal("FetchFoundation() error = nil, want unknown-command error")
	}
}
