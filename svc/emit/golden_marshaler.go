// golden_marshaler.go — converts emitted ampy protobufs into stable `Golden*` JSON structs for snapshot testing. Capacity: 3 `Golden*` types (bar, quote, fundamentals) + 3 `ToGolden*` converters + 1 `MarshalToGoldenJSON` wrapper that runs through `CanonicalMarshaler`.

package emit

import (
	"time"

	barsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/bars/v1"
	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	ticksv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/ticks/v1"
	"github.com/bizshuk/yfin/model"
)

// GoldenBarBatch represents the expected golden format for bar batches
type GoldenBarBatch struct {
	Security model.Security `json:"security"`
	Bars     []GoldenBar    `json:"bars"`
	Meta     model.Meta     `json:"meta"`
}

// GoldenBar represents the expected golden format for bars
type GoldenBar struct {
	Start              time.Time           `json:"start"`
	End                time.Time           `json:"end"`
	Open               model.ScaledDecimal `json:"open"`
	High               model.ScaledDecimal `json:"high"`
	Low                model.ScaledDecimal `json:"low"`
	Close              model.ScaledDecimal `json:"close"`
	Volume             int64               `json:"volume"`
	Adjusted           bool                `json:"adjusted"`
	AdjustmentPolicyID string              `json:"adjustment_policy_id"`
	EventTime          time.Time           `json:"event_time"`
	IngestTime         time.Time           `json:"ingest_time"`
	AsOf               time.Time           `json:"as_of"`
}

// GoldenQuote represents the expected golden format for quotes
type GoldenQuote struct {
	Security   model.Security       `json:"security"`
	Type       string               `json:"type"`
	Bid        *model.ScaledDecimal `json:"bid,omitempty"`
	BidSize    *int64               `json:"bid_size,omitempty"`
	Ask        *model.ScaledDecimal `json:"ask,omitempty"`
	AskSize    *int64               `json:"ask_size,omitempty"`
	Venue      string               `json:"venue"`
	EventTime  time.Time            `json:"event_time"`
	IngestTime time.Time            `json:"ingest_time"`
	Meta       model.Meta           `json:"meta"`
}

// GoldenFundamentals represents the expected golden format for fundamentals
type GoldenFundamentals struct {
	Security model.Security                     `json:"security"`
	Lines    []model.NormalizedFundamentalsLine `json:"lines"`
	Source   string                             `json:"source"`
	AsOf     time.Time                          `json:"as_of"`
	Meta     model.Meta                         `json:"meta"`
}

// ToGoldenBarBatch converts ampy-proto BarBatch to golden format
func ToGoldenBarBatch(barBatch *barsv1.BarBatch, security model.Security, meta model.Meta) *GoldenBarBatch {
	goldenBars := make([]GoldenBar, 0, len(barBatch.Bars))

	for _, bar := range barBatch.Bars {
		goldenBar := GoldenBar{
			Start:              bar.Start.AsTime(),
			End:                bar.End.AsTime(),
			Open:               model.ScaledDecimal{Scaled: bar.Open.Scaled, Scale: int(bar.Open.Scale)},
			High:               model.ScaledDecimal{Scaled: bar.High.Scaled, Scale: int(bar.High.Scale)},
			Low:                model.ScaledDecimal{Scaled: bar.Low.Scaled, Scale: int(bar.Low.Scale)},
			Close:              model.ScaledDecimal{Scaled: bar.Close.Scaled, Scale: int(bar.Close.Scale)},
			Volume:             bar.Volume,
			Adjusted:           bar.Adjusted,
			AdjustmentPolicyID: bar.AdjustmentPolicyId,
			EventTime:          bar.EventTime.AsTime(),
			IngestTime:         bar.IngestTime.AsTime(),
			AsOf:               bar.AsOf.AsTime(),
		}
		goldenBars = append(goldenBars, goldenBar)
	}

	return &GoldenBarBatch{
		Security: security,
		Bars:     goldenBars,
		Meta:     meta,
	}
}

// ToGoldenQuote converts ampy-proto QuoteTick to golden format
func ToGoldenQuote(quote *ticksv1.QuoteTick, meta model.Meta) *GoldenQuote {
	var bid, ask *model.ScaledDecimal
	var bidSize, askSize *int64

	if quote.Bid != nil {
		bid = &model.ScaledDecimal{Scaled: quote.Bid.Scaled, Scale: int(quote.Bid.Scale)}
	}
	if quote.Ask != nil {
		ask = &model.ScaledDecimal{Scaled: quote.Ask.Scaled, Scale: int(quote.Ask.Scale)}
	}
	if quote.BidSize != 0 {
		bidSize = &quote.BidSize
	}
	if quote.AskSize != 0 {
		askSize = &quote.AskSize
	}

	return &GoldenQuote{
		Security:   model.Security{Symbol: quote.Security.Symbol, MIC: quote.Security.Mic},
		Type:       "QUOTE",
		Bid:        bid,
		BidSize:    bidSize,
		Ask:        ask,
		AskSize:    askSize,
		Venue:      quote.Venue,
		EventTime:  quote.EventTime.AsTime(),
		IngestTime: quote.IngestTime.AsTime(),
		Meta:       meta,
	}
}

// ToGoldenFundamentals converts ampy-proto FundamentalsSnapshot to golden format
func ToGoldenFundamentals(fundamentals *fundamentalsv1.FundamentalsSnapshot, meta model.Meta) *GoldenFundamentals {
	goldenLines := make([]model.NormalizedFundamentalsLine, 0, len(fundamentals.Lines))

	for _, line := range fundamentals.Lines {
		goldenLine := model.NormalizedFundamentalsLine{
			Key:          line.Key,
			Value:        model.ScaledDecimal{Scaled: line.Value.Scaled, Scale: int(line.Value.Scale)},
			CurrencyCode: line.CurrencyCode,
			PeriodStart:  line.PeriodStart.AsTime(),
			PeriodEnd:    line.PeriodEnd.AsTime(),
		}
		goldenLines = append(goldenLines, goldenLine)
	}

	return &GoldenFundamentals{
		Security: model.Security{Symbol: fundamentals.Security.Symbol, MIC: fundamentals.Security.Mic},
		Lines:    goldenLines,
		Source:   fundamentals.Source,
		AsOf:     fundamentals.AsOf.AsTime(),
		Meta:     meta,
	}
}

// MarshalToGoldenJSON marshals to the expected golden JSON format
func MarshalToGoldenJSON(v interface{}) ([]byte, error) {
	return CanonicalMarshaler.Marshal(v)
}
