// bars.go — type aliases (`Bar`, `BarBatch`) for back-compat with the
// pre-model/ public surface + `FromBarBatch` ScaledDecimal → float64
// converter + `formatUTC` helper.
//
// The structs themselves now live in `github.com/bizshuk/yfin/model/bars.go`;
// `facade.Bar` and `facade.BarBatch` are type aliases (`type X = model.X`),
// so existing callers that still write `facade.Bar` keep compiling unchanged.
// New code should prefer `model.Bar` directly to avoid the indirection.

// Package facade re-exports the yfinance-go normalized bar/quote/company-info
// types as plain Go structs so external consumers (e.g. data/stock) can avoid
// reflection against the internal/norm package. The facade lives inside the
// yfinance-go module so it can legally import the internal types; consumers
// stay clean of `internal/`.
package facade

import (
	"time"

	"github.com/bizshuk/yfin/model"
)

// Bar is one daily OHLCV bar with float64 prices (decoded from ScaledDecimal).
// Aliased from model.Bar — new code should use model.Bar directly.
type Bar = model.Bar

// BarBatch is a series of Bars for one symbol, plus identifying metadata.
// Aliased from model.BarBatch — new code should use model.BarBatch directly.
type BarBatch = model.BarBatch

// FromBarBatch converts the internal model batch into a clean public struct.
// Returns nil if b is nil so callers can treat nil-checks uniformly.
func FromBarBatch(b *model.NormalizedBarBatch) *BarBatch {
	if b == nil {
		return nil
	}
	out := &BarBatch{
		Symbol: b.Security.Symbol,
		MIC:    b.Security.MIC,
		Bars:   make([]Bar, 0, len(b.Bars)),
	}
	for _, nb := range b.Bars {
		out.Bars = append(out.Bars, Bar{
			Date:         nb.EventTime.UTC().Format("2006-01-02"),
			Open:         model.FromScaledDecimal(nb.Open),
			High:         model.FromScaledDecimal(nb.High),
			Low:          model.FromScaledDecimal(nb.Low),
			Close:        model.FromScaledDecimal(nb.Close),
			Volume:       nb.Volume,
			Adjusted:     nb.Adjusted,
			CurrencyCode: nb.CurrencyCode,
		})
	}
	return out
}

// helper to expose time formatting in tests if needed elsewhere.
func formatUTC(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

var _ = formatUTC // reserved for future use by other facade files
