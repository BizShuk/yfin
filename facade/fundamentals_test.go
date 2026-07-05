package facade

import (
	"testing"
	"time"

	fundamentalsv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/fundamentals/v1"
	commonv1 "github.com/AmpyFin/ampy-proto/v2/gen/go/ampy/common/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestFromProtoFundamentals_Nil(t *testing.T) {
	assert.Nil(t, fromProtoFundamentals(nil), "nil proto must yield nil output")
}

func TestFromProtoFundamentals_TableDriven(t *testing.T) {
	startA := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endA := time.Date(2025, 3, 31, 23, 59, 59, 0, time.UTC)
	startB := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endB := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	asOf := time.Date(2025, 5, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   *fundamentalsv1.FundamentalsSnapshot
		want FundamentalsSnapshot
	}{
		{
			name: "populated snapshot, two lines preserve order",
			in: &fundamentalsv1.FundamentalsSnapshot{
				Security: &commonv1.SecurityId{Symbol: "AAPL", Mic: "XNAS"},
				Lines: []*fundamentalsv1.LineItem{
					{
						Key:          "revenue",
						Value:        &commonv1.Decimal{Scaled: 1234567, Scale: 2}, // 12345.67
						CurrencyCode: "USD",
						PeriodStart:  timestamppb.New(startA),
						PeriodEnd:    timestamppb.New(endA),
					},
					{
						Key:          "net_income",
						Value:        &commonv1.Decimal{Scaled: 9876543, Scale: 2}, // 98765.43
						CurrencyCode: "USD",
						PeriodStart:  timestamppb.New(startB),
						PeriodEnd:    timestamppb.New(endB),
					},
				},
				Source: "yfinance",
				AsOf:   timestamppb.New(asOf),
			},
			want: FundamentalsSnapshot{
				Symbol: "AAPL",
				MIC:    "XNAS",
				Source: "yfinance",
				AsOf:   asOf,
				Lines: []FundamentalsLine{
					{
						Key:          "revenue",
						Value:        12345.67,
						CurrencyCode: "USD",
						PeriodStart:  startA,
						PeriodEnd:    endA,
					},
					{
						Key:          "net_income",
						Value:        98765.43,
						CurrencyCode: "USD",
						PeriodStart:  startB,
						PeriodEnd:    endB,
					},
				},
			},
		},
		{
			name: "nil Security falls back to zero-value symbol/mic",
			in: &fundamentalsv1.FundamentalsSnapshot{
				Lines:  nil,
				Source: "tiingo",
				AsOf:   timestamppb.New(asOf),
			},
			want: FundamentalsSnapshot{
				Symbol: "",
				MIC:    "",
				Source: "tiingo",
				AsOf:   asOf,
				Lines:  []FundamentalsLine{},
			},
		},
		{
			name: "nil line Value surfaces as 0.0",
			in: &fundamentalsv1.FundamentalsSnapshot{
				Security: &commonv1.SecurityId{Symbol: "MSFT", Mic: "XNAS"},
				Lines: []*fundamentalsv1.LineItem{
					{
						Key:          "eps_basic",
						Value:        nil, // missing -> 0.0
						CurrencyCode: "USD",
						PeriodStart:  timestamppb.New(startA),
						PeriodEnd:    timestamppb.New(endA),
					},
				},
				Source: "yfinance",
				AsOf:   timestamppb.New(asOf),
			},
			want: FundamentalsSnapshot{
				Symbol: "MSFT",
				MIC:    "XNAS",
				Source: "yfinance",
				AsOf:   asOf,
				Lines: []FundamentalsLine{
					{
						Key:          "eps_basic",
						Value:        0.0,
						CurrencyCode: "USD",
						PeriodStart:  startA,
						PeriodEnd:    endA,
					},
				},
			},
		},
		{
			name: "nil AsOf yields zero time.Time",
			in: &fundamentalsv1.FundamentalsSnapshot{
				Security: &commonv1.SecurityId{Symbol: "GOOG", Mic: "XNAS"},
				Lines:    nil,
				Source:   "yfinance",
				AsOf:     nil,
			},
			want: FundamentalsSnapshot{
				Symbol: "GOOG",
				MIC:    "XNAS",
				Source: "yfinance",
				AsOf:   time.Time{},
				Lines:  []FundamentalsLine{},
			},
		},
		{
			name: "nil entry inside Lines slice is skipped, not zero-padded",
			in: &fundamentalsv1.FundamentalsSnapshot{
				Security: &commonv1.SecurityId{Symbol: "TSLA", Mic: "XNAS"},
				Lines: []*fundamentalsv1.LineItem{
					nil, // explicit nil in slice — must be skipped
					{
						Key:          "revenue",
						Value:        &commonv1.Decimal{Scaled: 1, Scale: 0}, // 1
						CurrencyCode: "USD",
						PeriodStart:  timestamppb.New(startA),
						PeriodEnd:    timestamppb.New(endA),
					},
				},
				Source: "yfinance",
				AsOf:   timestamppb.New(asOf),
			},
			want: FundamentalsSnapshot{
				Symbol: "TSLA",
				MIC:    "XNAS",
				Source: "yfinance",
				AsOf:   asOf,
				Lines: []FundamentalsLine{
					{
						Key:          "revenue",
						Value:        1.0,
						CurrencyCode: "USD",
						PeriodStart:  startA,
						PeriodEnd:    endA,
					},
				},
			},
		},
		{
			name: "non-USD currency and zero decimal scale",
			in: &fundamentalsv1.FundamentalsSnapshot{
				Security: &commonv1.SecurityId{Symbol: "7203", Mic: "XTKS"},
				Lines: []*fundamentalsv1.LineItem{
					{
						Key:          "shares_outstanding",
						Value:        &commonv1.Decimal{Scaled: 1000000000, Scale: 0},
						CurrencyCode: "JPY",
						PeriodStart:  timestamppb.New(startA),
						PeriodEnd:    timestamppb.New(endA),
					},
				},
				Source: "yfinance",
				AsOf:   timestamppb.New(asOf),
			},
			want: FundamentalsSnapshot{
				Symbol: "7203",
				MIC:    "XTKS",
				Source: "yfinance",
				AsOf:   asOf,
				Lines: []FundamentalsLine{
					{
						Key:          "shares_outstanding",
						Value:        1e9,
						CurrencyCode: "JPY",
						PeriodStart:  startA,
						PeriodEnd:    endA,
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := fromProtoFundamentals(tc.in)
			require.NotNil(t, got, "fromProtoFundamentals must never return nil for a non-nil input")
			assert.Equal(t, tc.want.Symbol, got.Symbol)
			assert.Equal(t, tc.want.MIC, got.MIC)
			assert.Equal(t, tc.want.Source, got.Source)
			assert.True(t, tc.want.AsOf.Equal(got.AsOf), "AsOf mismatch: want %v, got %v", tc.want.AsOf, got.AsOf)
			require.Equal(t, len(tc.want.Lines), len(got.Lines), "line count mismatch")
			for i, wantLine := range tc.want.Lines {
				gotLine := got.Lines[i]
				assert.Equal(t, wantLine.Key, gotLine.Key, "line[%d] key", i)
				assert.InDelta(t, wantLine.Value, gotLine.Value, 1e-9, "line[%d] value", i)
				assert.Equal(t, wantLine.CurrencyCode, gotLine.CurrencyCode, "line[%d] currency", i)
				assert.True(t, wantLine.PeriodStart.Equal(gotLine.PeriodStart), "line[%d] period_start", i)
				assert.True(t, wantLine.PeriodEnd.Equal(gotLine.PeriodEnd), "line[%d] period_end", i)
			}
		})
	}
}

// TestFromProtoFundamentals_OrderingIsPreserved documents the explicit design
// decision: ampy-proto uses []*LineItem (slice), so the converter preserves
// proto order without re-sorting. If a future caller needs sorted output,
// they should sort their own view rather than rely on this converter.
func TestFromProtoFundamentals_OrderingIsPreserved(t *testing.T) {
	keys := []string{"z_revenue", "a_net_income", "m_eps"}
	lines := make([]*fundamentalsv1.LineItem, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, &fundamentalsv1.LineItem{
			Key:          k,
			Value:        &commonv1.Decimal{Scaled: 1, Scale: 0},
			CurrencyCode: "USD",
		})
	}
	in := &fundamentalsv1.FundamentalsSnapshot{
		Security: &commonv1.SecurityId{Symbol: "NVDA", Mic: "XNAS"},
		Lines:    lines,
		Source:   "yfinance",
	}
	out := fromProtoFundamentals(in)
	require.NotNil(t, out)
	require.Equal(t, len(keys), len(out.Lines))
	for i, k := range keys {
		assert.Equal(t, k, out.Lines[i].Key, "ordering must be preserved at index %d", i)
	}
}