// bars.go — `/v8/finance/chart` HTTP fetch + JSON decode for the
// chart/bars endpoint. Type definitions (ChartResponse, Chart, ChartResult,
// ChartMeta, ChartIndicators, ChartBar, etc.) live in `model/yahoo_raw.go`;
// this file only owns the HTTP/JSON-decode behavior and exposes
// `model.*` types via type aliases for back-compat.
package yahoo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/bizshuk/yfin/model"
)

// Back-compat type aliases. New code should import `model` directly
// (e.g. `model.ChartResponse` / `model.ChartBar`). These aliases keep
// existing callers (`facade.Client.Fetch*`, dispatch wrappers) compiling
// without churn.
type (
	BarsResponse        = model.ChartResponse
	Chart               = model.Chart
	ChartResult         = model.ChartResult
	ChartMeta           = model.ChartMeta
	CurrentTradingPeriod = model.CurrentTradingPeriod
	TradingPeriod       = model.TradingPeriod
	ChartIndicators     = model.ChartIndicators
	QuoteIndicator      = model.QuoteIndicator
	AdjCloseIndicator   = model.AdjCloseIndicator
	Bar                 = model.ChartBar
)

// DecodeBarsResponse decodes a Yahoo Finance bars response (relaxed JSON:
// unknown fields allowed because Yahoo adds new fields frequently).
func DecodeBarsResponse(data []byte) (*model.ChartResponse, error) {
	var response model.ChartResponse
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode bars response: %w", err)
	}
	return &response, nil
}

// DecodeBarsResponseFromReader is the streaming variant of DecodeBarsResponse.
func DecodeBarsResponseFromReader(reader io.Reader) (*model.ChartResponse, error) {
	var response model.ChartResponse
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode bars response: %w", err)
	}
	return &response, nil
}

// ErrChartIncomplete is returned when a chart response is structurally
// valid but missing required OHLCV slices. Surfaced by svc-level
// validation functions (kept separate from the model-level ErrNoChartResults
// sentinel because validation is behavior, not data).
var ErrChartIncomplete = errors.New("chart response has no results")

// ValidateBars runs structural validation on a decoded chart response.
// Lives in svc/yahoo because validation is behavior; the data shape itself
// is owned by model.
func ValidateBars(r *model.ChartResponse) error {
	if r == nil || len(r.Chart.Result) == 0 {
		return ErrChartIncomplete
	}
	for i, result := range r.Chart.Result {
		if result.Meta.Symbol == "" {
			return fmt.Errorf("result[%d]: missing symbol", i)
		}
		if result.Meta.Currency == "" {
			return fmt.Errorf("result[%d]: missing currency", i)
		}
		if len(result.Timestamp) == 0 {
			continue
		}
		quote := result.Indicators.Quote[0]
		expectedLen := len(result.Timestamp)
		if len(quote.Open) != expectedLen ||
			len(quote.High) != expectedLen ||
			len(quote.Low) != expectedLen ||
			len(quote.Close) != expectedLen ||
			len(quote.Volume) != expectedLen {
			return fmt.Errorf("result[%d]: data length mismatch", i)
		}
		// Per-bar validation: skip bars with missing data, validate the rest.
		validBars := 0
		for j := 0; j < expectedLen; j++ {
			if quote.Open[j] == nil || quote.High[j] == nil || quote.Low[j] == nil || quote.Close[j] == nil || quote.Volume[j] == nil {
				continue
			}
			if err := validateBarData(quote.Open[j], quote.High[j], quote.Low[j], quote.Close[j], quote.Volume[j]); err != nil {
				return fmt.Errorf("result[%d].bar[%d]: %w", i, j, err)
			}
			validBars++
		}
		if validBars == 0 {
			return fmt.Errorf("result[%d]: no valid bars found", i)
		}
	}
	return nil
}

// validateBarData validates a single OHLCV bar's price + volume
// invariants and clamps high/low to ensure OHLC consistency.
func validateBarData(open, high, low, closePrice *float64, volume *int64) error {
	if open == nil || high == nil || low == nil || closePrice == nil || volume == nil {
		return fmt.Errorf("missing field")
	}
	if err := validatePrice(*open); err != nil {
		return fmt.Errorf("invalid open: %w", err)
	}
	if err := validatePrice(*high); err != nil {
		return fmt.Errorf("invalid high: %w", err)
	}
	if err := validatePrice(*low); err != nil {
		return fmt.Errorf("invalid low: %w", err)
	}
	if err := validatePrice(*closePrice); err != nil {
		return fmt.Errorf("invalid close: %w", err)
	}
	if *volume < 0 {
		return fmt.Errorf("negative volume: %d", *volume)
	}
	if *high < *low {
		*high = *low
	}
	if *high < *open {
		*high = *open
	}
	if *high < *closePrice {
		*high = *closePrice
	}
	if *low > *open {
		*low = *open
	}
	if *low > *closePrice {
		*low = *closePrice
	}
	return nil
}

// IsAdjusted reports whether the chart response carries adjusted-close data.
func IsAdjusted(r *model.ChartResponse) bool {
	if r == nil || len(r.Chart.Result) == 0 {
		return false
	}
	return len(r.Chart.Result[0].Indicators.AdjClose) > 0
}