// validation_test.go — table-driven tests for every validator in `validation.go`. Capacity: 6 test functions covering `ValidateSecurity`, `ValidateTimeWindow`, `ValidateDecimal`, `ValidateCurrency`, `ValidateAdjustments`, `ValidateFundamentals`.

package emit

import (
	"testing"
	"time"

	"github.com/bizshuk/yfin/model"
	"github.com/stretchr/testify/assert"
)

func TestValidateSecurity(t *testing.T) {
	tests := []struct {
		name     string
		security model.Security
		wantErr  bool
	}{
		{
			name:     "valid security",
			security: model.Security{Symbol: "AAPL", MIC: "XNAS"},
			wantErr:  false,
		},
		{
			name:     "empty symbol",
			security: model.Security{Symbol: "", MIC: "XNAS"},
			wantErr:  true,
		},
		{
			name:     "invalid MIC length",
			security: model.Security{Symbol: "AAPL", MIC: "XN"},
			wantErr:  true,
		},
		{
			name:     "invalid MIC format",
			security: model.Security{Symbol: "AAPL", MIC: "xnas"},
			wantErr:  true,
		},
		{
			name:     "valid without MIC",
			security: model.Security{Symbol: "AAPL"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecurity(tt.security)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTimeWindow(t *testing.T) {
	start := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	validEnd := start.Add(24 * time.Hour)
	validEvent := validEnd

	tests := []struct {
		name    string
		start   time.Time
		end     time.Time
		event   time.Time
		wantErr bool
	}{
		{
			name:    "valid daily window",
			start:   start,
			end:     validEnd,
			event:   validEvent,
			wantErr: false,
		},
		{
			name:    "invalid end time",
			start:   start,
			end:     start.Add(25 * time.Hour),
			event:   validEvent,
			wantErr: true,
		},
		{
			name:    "invalid event time",
			start:   start,
			end:     validEnd,
			event:   start.Add(12 * time.Hour),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeWindow(tt.start, tt.end, tt.event)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDecimal(t *testing.T) {
	tests := []struct {
		name    string
		decimal model.ScaledDecimal
		wantErr bool
	}{
		{
			name:    "valid decimal",
			decimal: model.ScaledDecimal{Scaled: 12345, Scale: 2},
			wantErr: false,
		},
		{
			name:    "negative scale",
			decimal: model.ScaledDecimal{Scaled: 12345, Scale: -1},
			wantErr: true,
		},
		{
			name:    "scale too large",
			decimal: model.ScaledDecimal{Scaled: 12345, Scale: 10},
			wantErr: true,
		},
		{
			name:    "zero scale",
			decimal: model.ScaledDecimal{Scaled: 12345, Scale: 0},
			wantErr: false,
		},
		{
			name:    "max scale",
			decimal: model.ScaledDecimal{Scaled: 12345, Scale: 9},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDecimal(tt.decimal)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCurrency(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "valid USD",
			code:    "USD",
			wantErr: false,
		},
		{
			name:    "valid EUR",
			code:    "EUR",
			wantErr: false,
		},
		{
			name:    "empty currency",
			code:    "",
			wantErr: true,
		},
		{
			name:    "wrong length",
			code:    "US",
			wantErr: true,
		},
		{
			name:    "lowercase",
			code:    "usd",
			wantErr: true,
		},
		{
			name:    "unknown currency (pass-through)",
			code:    "XYZ",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrency(tt.code)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAdjustments(t *testing.T) {
	tests := []struct {
		name     string
		adjusted bool
		policy   string
		wantErr  bool
	}{
		{
			name:     "valid raw",
			adjusted: false,
			policy:   "raw",
			wantErr:  false,
		},
		{
			name:     "valid split_only",
			adjusted: true,
			policy:   "split_only",
			wantErr:  false,
		},
		{
			name:     "valid split_dividend",
			adjusted: true,
			policy:   "split_dividend",
			wantErr:  false,
		},
		{
			name:     "invalid policy",
			adjusted: true,
			policy:   "invalid",
			wantErr:  true,
		},
		{
			name:     "inconsistent adjusted=true with raw",
			adjusted: true,
			policy:   "raw",
			wantErr:  true,
		},
		{
			name:     "inconsistent adjusted=false with split_only",
			adjusted: false,
			policy:   "split_only",
			wantErr:  true,
		},
		{
			name:     "inconsistent adjusted=false with split_dividend",
			adjusted: false,
			policy:   "split_dividend",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAdjustments(tt.adjusted, tt.policy)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFundamentals(t *testing.T) {
	validLine := model.NormalizedFundamentalsLine{
		Key:          "revenue",
		Value:        model.ScaledDecimal{Scaled: 1000000, Scale: 2},
		CurrencyCode: "USD",
		PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name    string
		lines   []model.NormalizedFundamentalsLine
		wantErr bool
	}{
		{
			name:    "valid fundamentals",
			lines:   []model.NormalizedFundamentalsLine{validLine},
			wantErr: false,
		},
		{
			name: "invalid key",
			lines: []model.NormalizedFundamentalsLine{
				{
					Key:          "invalid_key",
					Value:        model.ScaledDecimal{Scaled: 1000000, Scale: 2},
					CurrencyCode: "USD",
					PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					PeriodEnd:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
				},
			},
			wantErr: true,
		},
		{
			name: "custom key (valid)",
			lines: []model.NormalizedFundamentalsLine{
				{
					Key:          "custom_metric",
					Value:        model.ScaledDecimal{Scaled: 1000000, Scale: 2},
					CurrencyCode: "USD",
					PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					PeriodEnd:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
				},
			},
			wantErr: false,
		},
		{
			name: "invalid period",
			lines: []model.NormalizedFundamentalsLine{
				{
					Key:          "revenue",
					Value:        model.ScaledDecimal{Scaled: 1000000, Scale: 2},
					CurrencyCode: "USD",
					PeriodStart:  time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
					PeriodEnd:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFundamentals(tt.lines)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
