// market_data_test.go — `FromMarketData` ScaledDecimal → `*float64` round-trip
// tests: 1 nil-input case + 4 table-driven cases (full-population precision
// round-trip, all-nil nullable fields, scale-4 crypto/FX, zero-volume non-nil
// pointer). Capacity: 5 sub-tests + 3 helpers (`int64Ptr`, `ptr`,
// `checkFloatPtr`).
package facade

import (
	"testing"
	"time"

	"github.com/bizshuk/yfin/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromMarketData_Nil(t *testing.T) {
	assert.Nil(t, FromMarketData(nil), "nil input must yield nil output")
}

// int64Ptr returns a pointer to the given int64; small helper so the
// table cases stay readable.
func int64Ptr(v int64) *int64 { return &v }

func TestFromMarketData_TableDriven(t *testing.T) {
	eventTime := time.Date(2025, 5, 10, 15, 30, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   *model.NormalizedMarketData
		want MarketData
	}{
		{
			name: "all ScaledDecimals populated — full precision round-trip",
			in: &model.NormalizedMarketData{
				Security: model.Security{Symbol: "AAPL", MIC: "XNAS"},
				// Prices at scale 2: 12345.67, 12400.00, 12300.50, etc.
				RegularMarketPrice:  &model.ScaledDecimal{Scaled: 1234567, Scale: 2},
				RegularMarketHigh:   &model.ScaledDecimal{Scaled: 1240000, Scale: 2},
				RegularMarketLow:    &model.ScaledDecimal{Scaled: 1230050, Scale: 2},
				RegularMarketVolume: int64Ptr(50_000_000),
				FiftyTwoWeekHigh:    &model.ScaledDecimal{Scaled: 1990000, Scale: 2},
				FiftyTwoWeekLow:     &model.ScaledDecimal{Scaled: 1640000, Scale: 2},
				PreviousClose:       &model.ScaledDecimal{Scaled: 1233000, Scale: 2},
				CurrencyCode:        "USD",
				EventTime:           eventTime,
			},
			want: MarketData{
				Symbol:              "AAPL",
				MIC:                 "XNAS",
				RegularMarketPrice:  ptr(12345.67),
				RegularMarketHigh:   ptr(12400.00),
				RegularMarketLow:    ptr(12300.50),
				RegularMarketVolume: int64Ptr(50_000_000),
				FiftyTwoWeekHigh:    ptr(19900.00),
				FiftyTwoWeekLow:     ptr(16400.00),
				PreviousClose:       ptr(12330.00),
				CurrencyCode:        "USD",
				EventTime:           eventTime,
			},
		},
		{
			name: "all nullable fields nil — nil pointers in output",
			in: &model.NormalizedMarketData{
				Security:            model.Security{Symbol: "TSLA", MIC: "XNAS"},
				RegularMarketPrice:  nil,
				RegularMarketHigh:   nil,
				RegularMarketLow:    nil,
				RegularMarketVolume: nil,
				FiftyTwoWeekHigh:    nil,
				FiftyTwoWeekLow:     nil,
				PreviousClose:       nil,
				CurrencyCode:        "USD",
				EventTime:           eventTime,
			},
			want: MarketData{
				Symbol:       "TSLA",
				MIC:          "XNAS",
				CurrencyCode: "USD",
				EventTime:    eventTime,
				// All *float64 / *int64 stay nil — verified via .IsNil() below.
			},
		},
		{
			name: "non-USD currency, scale 4 (e.g., crypto or FX)",
			in: &model.NormalizedMarketData{
				Security: model.Security{Symbol: "BTC-USD", MIC: "XNAS"},
				// 6543.2101 at scale 4
				RegularMarketPrice: &model.ScaledDecimal{Scaled: 65432101, Scale: 4},
				RegularMarketHigh:  &model.ScaledDecimal{Scaled: 66000000, Scale: 4},
				RegularMarketLow:   &model.ScaledDecimal{Scaled: 65000000, Scale: 4},
				FiftyTwoWeekHigh:   &model.ScaledDecimal{Scaled: 73000000, Scale: 4},
				FiftyTwoWeekLow:    &model.ScaledDecimal{Scaled: 25000000, Scale: 4},
				PreviousClose:      &model.ScaledDecimal{Scaled: 65100000, Scale: 4},
				CurrencyCode:       "USD",
				EventTime:          eventTime,
			},
			want: MarketData{
				Symbol:             "BTC-USD",
				MIC:                "XNAS",
				RegularMarketPrice: ptr(6543.2101),
				RegularMarketHigh:  ptr(6600.0000),
				RegularMarketLow:   ptr(6500.0000),
				FiftyTwoWeekHigh:   ptr(7300.0000),
				FiftyTwoWeekLow:    ptr(2500.0000),
				PreviousClose:      ptr(6510.0000),
				CurrencyCode:       "USD",
				EventTime:          eventTime,
			},
		},
		{
			name: "zero-decimal volume but non-nil pointer",
			in: &model.NormalizedMarketData{
				Security:            model.Security{Symbol: "INDEX", MIC: "XNAS"},
				RegularMarketPrice:  &model.ScaledDecimal{Scaled: 500000, Scale: 2}, // 5000.00
				RegularMarketVolume: int64Ptr(0),                                    // explicitly zero
				CurrencyCode:        "USD",
				EventTime:           eventTime,
			},
			want: MarketData{
				Symbol:              "INDEX",
				MIC:                 "XNAS",
				RegularMarketPrice:  ptr(5000.00),
				RegularMarketVolume: int64Ptr(0), // not nil — the value is 0, not missing
				CurrencyCode:        "USD",
				EventTime:           eventTime,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FromMarketData(tc.in)
			require.NotNil(t, got, "FromMarketData must never return nil for a non-nil input")

			assert.Equal(t, tc.want.Symbol, got.Symbol)
			assert.Equal(t, tc.want.MIC, got.MIC)
			assert.Equal(t, tc.want.CurrencyCode, got.CurrencyCode)
			assert.True(t, tc.want.EventTime.Equal(got.EventTime),
				"EventTime mismatch: want %v, got %v", tc.want.EventTime, got.EventTime)

			// Nullable *float64 fields.
			checkFloatPtr(t, "RegularMarketPrice", tc.want.RegularMarketPrice, got.RegularMarketPrice)
			checkFloatPtr(t, "RegularMarketHigh", tc.want.RegularMarketHigh, got.RegularMarketHigh)
			checkFloatPtr(t, "RegularMarketLow", tc.want.RegularMarketLow, got.RegularMarketLow)
			checkFloatPtr(t, "FiftyTwoWeekHigh", tc.want.FiftyTwoWeekHigh, got.FiftyTwoWeekHigh)
			checkFloatPtr(t, "FiftyTwoWeekLow", tc.want.FiftyTwoWeekLow, got.FiftyTwoWeekLow)
			checkFloatPtr(t, "PreviousClose", tc.want.PreviousClose, got.PreviousClose)

			// Nullable *int64 field.
			if tc.want.RegularMarketVolume == nil {
				assert.Nil(t, got.RegularMarketVolume, "RegularMarketVolume should be nil")
			} else {
				require.NotNil(t, got.RegularMarketVolume, "RegularMarketVolume should not be nil")
				assert.Equal(t, *tc.want.RegularMarketVolume, *got.RegularMarketVolume)
			}
		})
	}
}

// ptr is a generic helper for taking the address of a literal in a struct
// literal — needed for *float64 fields in the want value.
func ptr(v float64) *float64 { return &v }

// checkFloatPtr compares a nullable float64 pointer pair. nil == nil is OK;
// otherwise both must be non-nil and within delta.
func checkFloatPtr(t *testing.T, name string, want, got *float64) {
	t.Helper()
	if want == nil {
		assert.Nil(t, got, "%s should be nil", name)
		return
	}
	require.NotNil(t, got, "%s should not be nil", name)
	assert.InDelta(t, *want, *got, 1e-9, "%s mismatch: want %v, got %v", name, *want, *got)
}
