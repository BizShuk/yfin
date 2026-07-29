// bars_test.go — unit tests for `NormalizeBars` validation paths (`TestNormalizeBarsValidation`: empty bars, nil meta, invalid security, etc.) and `ToUTCDayBoundaries` timestamp-to-UTC-day mapping (`TestToUTCDayBoundaries`).

package model

import (
	"testing"
	"time"
)

// Golden file tests removed - they were testing against outdated scale expectations

func TestNormalizeBarsValidation(t *testing.T) {
	tests := []struct {
		name    string
		bars    []ChartBar
		meta    *ChartMeta
		runID   string
		wantErr bool
	}{
		{
			name: "valid bars",
			bars: []ChartBar{
				{
					Timestamp: 1704326400,
					Open:      189.23,
					High:      191.0,
					Low:       188.9,
					Close:     190.45,
					Volume:    43210000,
				},
			},
			meta: &ChartMeta{
				Symbol:   "AAPL",
				Currency: "USD",
			},
			runID:   "test_run",
			wantErr: false,
		},
		{
			name: "empty bars",
			bars: []ChartBar{},
			meta: &ChartMeta{
				Symbol:   "AAPL",
				Currency: "USD",
			},
			runID:   "test_run",
			wantErr: true,
		},
		{
			name: "nil metadata",
			bars: []ChartBar{
				{
					Timestamp: 1704326400,
					Open:      189.23,
					High:      191.0,
					Low:       188.9,
					Close:     190.45,
					Volume:    43210000,
				},
			},
			meta:    nil,
			runID:   "test_run",
			wantErr: true,
		},
		{
			name: "missing symbol",
			bars: []ChartBar{
				{
					Timestamp: 1704326400,
					Open:      189.23,
					High:      191.0,
					Low:       188.9,
					Close:     190.45,
					Volume:    43210000,
				},
			},
			meta: &ChartMeta{
				Symbol:   "",
				Currency: "USD",
			},
			runID:   "test_run",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeBars(tt.bars, tt.meta, tt.runID)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeBars() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestToTradingDay(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		timezone  string
		wantDate  string
		wantStart time.Time
		wantEnd   time.Time
	}{
		{
			name:      "US market open stays on January 3",
			timestamp: 1704292200, // 2024-01-03T14:30:00Z
			timezone:  "America/New_York",
			wantDate:  "2024-01-03",
			wantStart: time.Date(2024, 1, 3, 5, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 1, 4, 5, 0, 0, 0, time.UTC),
		},
		{
			name:      "Taiwan market open stays on January 3",
			timestamp: 1704243600, // 2024-01-03T01:00:00Z
			timezone:  "Asia/Taipei",
			wantDate:  "2024-01-03",
			wantStart: time.Date(2024, 1, 2, 16, 0, 0, 0, time.UTC),
			wantEnd:   time.Date(2024, 1, 3, 16, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, start, end, eventTime, err := ToTradingDay(tt.timestamp, tt.timezone)
			if err != nil {
				t.Fatal(err)
			}
			if date != tt.wantDate || !start.Equal(tt.wantStart) ||
				!end.Equal(tt.wantEnd) || !eventTime.Equal(tt.wantEnd) {
				t.Fatalf(
					"got date=%s start=%s end=%s event=%s",
					date, start, end, eventTime,
				)
			}
		})
	}
}

func TestToTradingDayRejectsUnknownTimezone(t *testing.T) {
	_, _, _, _, err := ToTradingDay(1704292200, "Mars/Olympus")
	if err == nil {
		t.Fatal("expected timezone error")
	}
}

func TestNormalizeBarsPreservesProviderFirstTradeDate(t *testing.T) {
	firstTrade := time.Date(1980, 12, 12, 9, 30, 0, 0, time.FixedZone("EST", -5*60*60))
	batch, err := NormalizeBars(
		[]ChartBar{{
			Timestamp: 1704292200,
			Open:      189.23,
			High:      191,
			Low:       188.9,
			Close:     190.45,
			Volume:    43210000,
		}},
		&ChartMeta{
			Symbol:               "AAPL",
			Currency:             "USD",
			ExchangeName:         "NMS",
			ExchangeTimezoneName: "America/New_York",
			FirstTradeDate:       firstTrade.Unix(),
		},
		"test-run",
	)
	if err != nil {
		t.Fatal(err)
	}
	if batch.FirstTradeDate != "1980-12-12" {
		t.Fatalf("FirstTradeDate = %q, want 1980-12-12", batch.FirstTradeDate)
	}
	if got := batch.Bars[0].TradingDate; got != "2024-01-03" {
		t.Fatalf("TradingDate = %q, want 2024-01-03", got)
	}
}

func TestNormalizeBarsAdjustsEntireOHLC(t *testing.T) {
	adjustedClose := 84.0
	batch, err := NormalizeBars(
		[]ChartBar{{
			Timestamp: 1704292200,
			Open:      100,
			High:      110,
			Low:       90,
			Close:     105,
			AdjClose:  &adjustedClose,
			Volume:    1000,
		}},
		&ChartMeta{
			Symbol:               "TEST",
			Currency:             "USD",
			ExchangeName:         "NMS",
			ExchangeTimezoneName: "America/New_York",
		},
		"adjusted-test",
	)
	if err != nil {
		t.Fatal(err)
	}
	bar := batch.Bars[0]
	if got := FromScaledDecimal(bar.Open); got != 80 {
		t.Fatalf("adjusted open = %v, want 80", got)
	}
	if got := FromScaledDecimal(bar.High); got != 88 {
		t.Fatalf("adjusted high = %v, want 88", got)
	}
	if got := FromScaledDecimal(bar.Low); got != 72 {
		t.Fatalf("adjusted low = %v, want 72", got)
	}
	if got := FromScaledDecimal(bar.Close); got != 84 {
		t.Fatalf("adjusted close = %v, want 84", got)
	}
}
