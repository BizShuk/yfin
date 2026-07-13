// fundamentals.go — `FundamentalsLine` / `FundamentalsSnapshot` type aliases
// + the public `FromFundamentalsSnapshot` (norm → SDK) converter. Structs
// live in model/fundamentals.go; facade re-exports them as back-compat
// aliases. The former `fromProtoFundamentals` proto→SDK converter was removed
// with the ampy-proto dependency — scrape DTOs now convert to model directly
// via model.Scrape*ToSnapshot.
package facade

import (
	"github.com/bizshuk/yfin/model"
)

// FundamentalsLine is one period-tagged line item. Aliased from
// model.FundamentalsLine — new code should use model.FundamentalsLine directly.
type FundamentalsLine = model.FundamentalsLine

// FundamentalsSnapshot is one symbol's fundamentals view. Aliased from
// model.FundamentalsSnapshot — new code should use model.FundamentalsSnapshot
// directly.
type FundamentalsSnapshot = model.FundamentalsSnapshot

// FromFundamentalsSnapshot converts the internal model.NormalizedFundamentalsSnapshot
// into the plain SDK FundamentalsSnapshot. Uses model.FromScaledDecimal to
// unwrap each line's value; nil input -> nil output.
func FromFundamentalsSnapshot(n *model.NormalizedFundamentalsSnapshot) *FundamentalsSnapshot {
	if n == nil {
		return nil
	}
	out := &FundamentalsSnapshot{
		Symbol: n.Security.Symbol,
		MIC:    n.Security.MIC,
		Source: n.Source,
		AsOf:   n.AsOf,
		Lines:  make([]FundamentalsLine, 0, len(n.Lines)),
	}
	for _, line := range n.Lines {
		out.Lines = append(out.Lines, FundamentalsLine{
			Key:          line.Key,
			Value:        model.FromScaledDecimal(line.Value),
			CurrencyCode: line.CurrencyCode,
			PeriodStart:  line.PeriodStart,
			PeriodEnd:    line.PeriodEnd,
		})
	}
	return out
}
