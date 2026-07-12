// fundamentals.go — `FundamentalsLine` / `FundamentalsSnapshot` type aliases
// + 2 converters: public `FromFundamentalsSnapshot` (norm → SDK) and
// unexported `fromProtoFundamentals` (ampy-proto → SDK, used by
// `Client.Scrape*` only). Structs live in model/fundamentals.go; facade
// re-exports them as back-compat aliases.
package facade

import (
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
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

// fromProtoFundamentals converts an ampy-proto FundamentalsSnapshot into the
// plain SDK FundamentalsSnapshot. It is unexported because it is only used
// internally by facade.Client (Step 6) when handing scrape/proto data to
// SDK callers; the public entry point for SDK consumers is the
// From* family of converters.
//
// Notes on the conversion:
//   - Security: ampy-proto uses *commonv1.SecurityId with Symbol/Mic as
//     top-level fields (vs model.Security where MIC is also top-level); both
//     map 1:1.
//   - Value: the proto's *commonv1.Decimal is decoded to float64 via
//     model.FromScaledDecimal. A nil Decimal surfaces as 0 (float64 cannot
//     represent nil); CurrencyCode / Period* are copied verbatim.
//   - AsOf: ampy-proto Timestamp → time.Time (UTC). A nil AsOf yields the
//     zero time.Time.
//   - Lines ordering: ampy-proto uses []*LineItem (a slice), so we copy
//     lines in input order — no sort, no flatten.
func fromProtoFundamentals(p *fundamentalsv1.FundamentalsSnapshot) *FundamentalsSnapshot {
	if p == nil {
		return nil
	}
	out := &FundamentalsSnapshot{
		Source: p.Source,
		Lines:  make([]FundamentalsLine, 0, len(p.GetLines())),
	}
	if sec := p.GetSecurity(); sec != nil {
		out.Symbol = sec.GetSymbol()
		out.MIC = sec.GetMic()
	}
	if ts := p.GetAsOf(); ts != nil {
		out.AsOf = ts.AsTime().UTC()
	}
	for _, li := range p.GetLines() {
		if li == nil {
			continue
		}
		line := FundamentalsLine{
			Key:          li.GetKey(),
			CurrencyCode: li.GetCurrencyCode(),
		}
		if v := li.GetValue(); v != nil {
			// Decode the scaled integer back to float64. The proto's
			// Scale is int32 but scales are small (typ. 0-8) so the
			// conversion to model.ScaledDecimal.Scale (int) is lossless.
			line.Value = model.FromScaledDecimal(model.ScaledDecimal{
				Scaled: v.GetScaled(),
				Scale:  int(v.GetScale()),
			})
		}
		if ts := li.GetPeriodStart(); ts != nil {
			line.PeriodStart = ts.AsTime().UTC()
		}
		if ts := li.GetPeriodEnd(); ts != nil {
			line.PeriodEnd = ts.AsTime().UTC()
		}
		out.Lines = append(out.Lines, line)
	}
	return out
}
