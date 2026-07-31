package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFundamentalSnapshotAllowsMissingCurrency(t *testing.T) {
	snapshot := validFundamentalSnapshot()
	snapshot.Lines[0].CurrencyCode = ""

	path := writeFundamentalSnapshot(t, snapshot)
	if !validateFundamentalSnapshot(path) {
		t.Fatal("validateFundamentalSnapshot rejected an absent currency")
	}
}

func TestValidateFundamentalSnapshotRejectsInvalidCurrency(t *testing.T) {
	snapshot := validFundamentalSnapshot()
	snapshot.Lines[0].CurrencyCode = "US"

	path := writeFundamentalSnapshot(t, snapshot)
	if validateFundamentalSnapshot(path) {
		t.Fatal("validateFundamentalSnapshot accepted an invalid currency")
	}
}

func validFundamentalSnapshot() FundamentalSnapshot {
	return FundamentalSnapshot{
		Security: Security{Symbol: "AAPL"},
		Lines: []FundamentalLine{
			{
				Key:          "revenue",
				CurrencyCode: "USD",
			},
		},
		Meta: Meta{Source: yfinSource},
	}
}

func writeFundamentalSnapshot(t *testing.T, snapshot FundamentalSnapshot) string {
	t.Helper()

	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	path := filepath.Join(t.TempDir(), "snapshot.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
	return path
}
