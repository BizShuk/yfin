// Tests bars decoding + validation: `TestDecodeBarsResponse` loads 5 fixture files (`AAPL`, `SAP EUR`, `TM JPY`, `TSLA split window`) and asserts OHLC + volume invariants on the first bar; `TestBarsResponseValidation` covers happy path + missing-symbol/currency + negative-volume error paths.
package yahoo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bizshuk/yfin/model"
	"github.com/bizshuk/yfin/utils/httpx"
	"github.com/stretchr/testify/require"
)

func TestDecodeBarsResponse(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "AAPL daily bars",
			filename: "AAPL_1d_sample.json",
			wantErr:  false,
		},
		{
			name:     "AAPL raw bars",
			filename: "AAPL_1d_raw_sample.json",
			wantErr:  false,
		},
		{
			name:     "SAP EUR bars",
			filename: "SAP_XETR_1d_eur.json",
			wantErr:  false,
		},
		{
			name:     "TM JPY bars",
			filename: "TM_XTKS_1d_jpy.json",
			wantErr:  false,
		},
		{
			name:     "TSLA split window",
			filename: "TSLA_1d_split_window.json",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test data
			data, err := os.ReadFile(filepath.Join("../../tests/testdata/source/yahoo/bars", tt.filename))
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			// Decode response
			response, err := DecodeBarsResponse(data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeBarsResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Validate response structure
				if response == nil {
					t.Fatal("Response is nil")
				}

				// Check that we have results
				if len(response.Chart.Result) == 0 {
					t.Fatal("No chart results found")
				}

				result := response.Chart.Result[0]

				// Validate metadata
				if result.Meta.Symbol == "" {
					t.Error("Missing symbol in metadata")
				}
				if result.Meta.Currency == "" {
					t.Error("Missing currency in metadata")
				}

				// Validate bars data
				bars, err := response.GetBars()
				if err != nil {
					t.Errorf("Failed to get bars: %v", err)
				}

				if len(bars) == 0 {
					t.Error("No bars found")
				}

				// Validate first bar
				bar := bars[0]
				if bar.Timestamp == 0 {
					t.Error("Invalid timestamp")
				}
				if bar.Open <= 0 {
					t.Error("Invalid open price")
				}
				if bar.High <= 0 {
					t.Error("Invalid high price")
				}
				if bar.Low <= 0 {
					t.Error("Invalid low price")
				}
				if bar.Close <= 0 {
					t.Error("Invalid close price")
				}
				if bar.Volume < 0 {
					t.Error("Invalid volume")
				}

				// Check OHLC relationships
				if bar.High < bar.Low {
					t.Error("High < Low")
				}
				if bar.High < bar.Open {
					t.Error("High < Open")
				}
				if bar.High < bar.Close {
					t.Error("High < Close")
				}
				if bar.Low > bar.Open {
					t.Error("Low > Open")
				}
				if bar.Low > bar.Close {
					t.Error("Low > Close")
				}
			}
		})
	}
}

func TestBarsResponseValidation(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name: "valid response",
			data: `{
				"chart": {
					"result": [{
						"meta": {
							"currency": "USD",
							"symbol": "AAPL",
							"exchangeName": "NASDAQ"
						},
						"timestamp": [1704326400],
						"indicators": {
							"quote": [{
								"open": [189.23],
								"high": [191.0],
								"low": [188.9],
								"close": [190.45],
								"volume": [43210000]
							}]
						}
					}],
					"error": null
				}
			}`,
			wantErr: false,
		},
		{
			name: "missing symbol",
			data: `{
				"chart": {
					"result": [{
						"meta": {
							"currency": "USD",
							"symbol": "",
							"exchangeName": "NASDAQ"
						},
						"timestamp": [1704326400],
						"indicators": {
							"quote": [{
								"open": [189.23],
								"high": [191.0],
								"low": [188.9],
								"close": [190.45],
								"volume": [43210000]
							}]
						}
					}],
					"error": null
				}
			}`,
			wantErr: true,
		},
		{
			name: "missing currency",
			data: `{
				"chart": {
					"result": [{
						"meta": {
							"currency": "",
							"symbol": "AAPL",
							"exchangeName": "NASDAQ"
						},
						"timestamp": [1704326400],
						"indicators": {
							"quote": [{
								"open": [189.23],
								"high": [191.0],
								"low": [188.9],
								"close": [190.45],
								"volume": [43210000]
							}]
						}
					}],
					"error": null
				}
			}`,
			wantErr: false,
		},
		{
			name: "invalid volume",
			data: `{
				"chart": {
					"result": [{
						"meta": {
							"currency": "USD",
							"symbol": "AAPL",
							"exchangeName": "NASDAQ"
						},
						"timestamp": [1704326400],
						"indicators": {
							"quote": [{
								"open": [189.23],
								"high": [191.0],
								"low": [188.9],
								"close": [190.45],
								"volume": [-43210000]
							}]
						}
					}],
					"error": null
				}
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response model.ChartResponse
			err := json.Unmarshal([]byte(tt.data), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal test data: %v", err)
			}

			err = ValidateBars(&response)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFetchDailyBarsAppliesProviderOHLCIntegrityPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		_, err := writer.Write([]byte(`{
			"chart": {
				"result": [{
					"meta": {
						"currency": "USD",
						"symbol": "AA",
						"exchangeName": "NYQ",
						"exchangeTimezoneName": "America/New_York"
					},
					"timestamp": [1785245400],
					"indicators": {
						"quote": [{
							"open": [43.59],
							"high": [44.62],
							"low": [43.72],
							"close": [44.36],
							"volume": [4595283]
						}],
						"adjclose": [{
							"adjclose": [44.36]
						}]
					}
				}],
				"error": null
			}
		}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	client := NewClient(
		httpx.NewClient(httpx.DefaultConfig()),
		server.URL,
	)

	response, err := client.FetchDailyBars(
		context.Background(),
		"AA",
		time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC),
		true,
	)
	require.NoError(t, err)
	bars, err := response.GetBars()
	require.NoError(t, err)
	require.Len(t, bars, 1)
	require.Equal(t, bars[0].Open, bars[0].Low)
	require.GreaterOrEqual(t, bars[0].High, bars[0].Close)
}

func TestFetchDailyBarsAllowsMissingCurrencyMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(
		writer http.ResponseWriter,
		_ *http.Request,
	) {
		_, err := writer.Write([]byte(`{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "FBYDP",
						"exchangeName": "NGM",
						"exchangeTimezoneName": "America/New_York"
					},
					"timestamp": [1785245400],
					"indicators": {
						"quote": [{
							"open": [19.9],
							"high": [20.0],
							"low": [19.9],
							"close": [20.0],
							"volume": [100]
						}],
						"adjclose": [{
							"adjclose": [20.0]
						}]
					}
				}],
				"error": null
			}
		}`))
		require.NoError(t, err)
	}))
	defer server.Close()
	client := NewClient(
		httpx.NewClient(httpx.DefaultConfig()),
		server.URL,
	)

	response, err := client.FetchDailyBars(
		context.Background(),
		"FBYDP",
		time.Date(2026, 7, 28, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC),
		true,
	)
	require.NoError(t, err)
	require.Equal(t, "", response.GetMetadata().Currency)
}
