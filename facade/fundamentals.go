package facade

import (
	"time"

	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	"github.com/bizshuk/yfinance-go/svc/norm"
)

// FundamentalsLine is one period-tagged line item (e.g., "revenue",
// "net_income", "eps_basic"). Value is the float64 decoded from the
// internal ScaledDecimal / ampy-proto Decimal via norm.FromScaledDecimal;
// a missing Value is surfaced as 0 (float64 cannot distinguish zero from
// missing). PeriodStart / PeriodEnd are UTC.
type FundamentalsLine struct {
	Key          string    `json:"key"`
	Value        float64   `json:"value"`
	CurrencyCode string    `json:"currency_code,omitempty"`
	PeriodStart  time.Time `json:"period_start"`
	PeriodEnd    time.Time `json:"period_end"`
}

// FundamentalsSnapshot is one symbol's fundamentals view, exposed as a plain
// SDK struct. Lines preserves the order it was supplied in (the ampy-proto
// FundamentalsSnapshot.Lines is a []*LineItem, not a map, so ordering is
// inherent; fromProtoFundamentals and FromFundamentalsSnapshot both copy
// lines in input order without re-sorting).
type FundamentalsSnapshot struct {
	Symbol string             `json:"symbol"`
	MIC    string             `json:"mic,omitempty"`
	Source string             `json:"source,omitempty"`
	AsOf   time.Time          `json:"as_of"`
	Lines  []FundamentalsLine `json:"lines,omitempty"`
}

// FromFundamentalsSnapshot converts the internal norm.NormalizedFundamentalsSnapshot
// into the plain SDK FundamentalsSnapshot. Uses norm.FromScaledDecimal to
// unwrap each line's value; nil input -> nil output.
func FromFundamentalsSnapshot(n *norm.NormalizedFundamentalsSnapshot) *FundamentalsSnapshot {
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
			Value:        norm.FromScaledDecimal(line.Value),
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
//     top-level fields (vs norm.Security where MIC is also top-level); both
//     map 1:1.
//   - Value: the proto's *commonv1.Decimal is decoded to float64 via
//     norm.FromScaledDecimal. A nil Decimal surfaces as 0 (float64 cannot
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
			// conversion to norm.ScaledDecimal.Scale (int) is lossless.
			line.Value = norm.FromScaledDecimal(norm.ScaledDecimal{
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