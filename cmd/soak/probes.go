// probes.go — compares Yahoo API info with scraped statistics, financials, and profile data.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/model"
)

// CorrectnessProbes handles validation of API vs scrape data consistency.
type CorrectnessProbes struct {
	client *facade.Client
	logger *slog.Logger
}

// ProbeResult represents the result of a correctness probe.
type ProbeResult struct {
	Ticker      string
	ProbeType   string
	Passed      bool
	Details     string
	APIValue    any
	ScrapeValue any
	Tolerance   float64
	Difference  float64
}

type probeDataset struct {
	info       map[string]any
	stats      *model.ComprehensiveKeyStatisticsDTO
	financials *model.ComprehensiveFinancialsDTO
	profile    *model.ComprehensiveProfileDTO
}

// NewCorrectnessProbes creates a new correctness probe instance.
func NewCorrectnessProbes(client *facade.Client, logger *slog.Logger) *CorrectnessProbes {
	return &CorrectnessProbes{client: client, logger: logger}
}

// ValidateTicker fetches each source once and runs all comparison probes.
func (probes *CorrectnessProbes) ValidateTicker(ctx context.Context, ticker string) error {
	runID := fmt.Sprintf("probe-%s-%d", ticker, time.Now().Unix())
	probes.logger.Debug("Starting correctness validation", "ticker", ticker)

	infoValue, err := probes.client.FetchFoundation(ctx, "info", ticker, runID)
	if err != nil {
		return fmt.Errorf("fetch API info: %w", err)
	}
	info, ok := infoValue.(map[string]any)
	if !ok {
		return fmt.Errorf("fetch API info: unexpected response type %T", infoValue)
	}
	stats, err := probes.client.ScrapeKeyStatisticsData(ctx, ticker, runID)
	if err != nil {
		return fmt.Errorf("scrape key statistics: %w", err)
	}
	financials, err := probes.client.ScrapeFinancialsData(ctx, ticker, runID)
	if err != nil {
		return fmt.Errorf("scrape financials: %w", err)
	}
	profile, err := probes.client.ScrapeProfileData(ctx, ticker, runID)
	if err != nil {
		return fmt.Errorf("scrape profile: %w", err)
	}

	dataset := probeDataset{
		info:       info,
		stats:      stats,
		financials: financials,
		profile:    profile,
	}
	results := []*ProbeResult{
		probeMarketCap(ticker, dataset),
		probePERatio(ticker, dataset),
		probeEmployeeCount(ticker, dataset),
		probeSector(ticker, dataset),
		probeCurrency(ticker, dataset),
	}

	var failures []string
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
			probes.logger.Debug(
				"Probe passed",
				"ticker", ticker,
				"probe_type", result.ProbeType,
				"details", result.Details,
			)
			continue
		}
		failures = append(failures, fmt.Sprintf("%s: %s", result.ProbeType, result.Details))
		probes.logger.Warn(
			"Probe failed",
			"ticker", ticker,
			"probe_type", result.ProbeType,
			"details", result.Details,
			"api_value", result.APIValue,
			"scrape_value", result.ScrapeValue,
		)
	}

	required := len(results)/2 + 1
	if passed >= required {
		probes.logger.Debug(
			"Overall validation passed",
			"ticker", ticker,
			"passed_probes", passed,
			"total_probes", len(results),
		)
		return nil
	}
	return fmt.Errorf("validation failed for %s: %s", ticker, strings.Join(failures, "; "))
}

func probeMarketCap(ticker string, dataset probeDataset) *ProbeResult {
	result := &ProbeResult{Ticker: ticker, ProbeType: "market_cap", Tolerance: 0.05}
	apiValue, apiOK := numericInfo(dataset.info, "marketCap")
	scrapeValue, scrapeOK := scaledValue(dataset.stats.Current.MarketCap)
	return comparePositive(result, apiValue, apiOK, scrapeValue, scrapeOK)
}

func probePERatio(ticker string, dataset probeDataset) *ProbeResult {
	result := &ProbeResult{Ticker: ticker, ProbeType: "pe_ratio", Tolerance: 0.10}
	apiValue, apiOK := numericInfo(dataset.info, "trailingPE")
	scrapeValue, scrapeOK := scaledValue(dataset.stats.Current.TrailingPE)
	if !apiOK || !scrapeOK {
		apiValue, apiOK = numericInfo(dataset.info, "forwardPE")
		scrapeValue, scrapeOK = scaledValue(dataset.stats.Current.ForwardPE)
	}
	return comparePositive(result, apiValue, apiOK, scrapeValue, scrapeOK)
}

func probeEmployeeCount(ticker string, dataset probeDataset) *ProbeResult {
	result := &ProbeResult{Ticker: ticker, ProbeType: "employee_count", Tolerance: 0.15}
	apiValue, apiOK := numericInfo(dataset.info, "fullTimeEmployees")
	scrapeOK := dataset.profile != nil && dataset.profile.FullTimeEmployees != nil
	scrapeValue := float64(0)
	if scrapeOK {
		scrapeValue = float64(*dataset.profile.FullTimeEmployees)
	}
	return comparePositive(result, apiValue, apiOK, scrapeValue, scrapeOK)
}

func probeSector(ticker string, dataset probeDataset) *ProbeResult {
	result := &ProbeResult{Ticker: ticker, ProbeType: "sector"}
	apiValue, apiOK := stringInfo(dataset.info, "sector")
	scrapeValue := ""
	if dataset.profile != nil {
		scrapeValue = dataset.profile.Sector
	}
	result.APIValue = apiValue
	result.ScrapeValue = scrapeValue
	if !apiOK || strings.TrimSpace(scrapeValue) == "" {
		result.Details = "sector is unavailable from one or both real sources"
		return result
	}
	result.Passed = normalizeSectorName(apiValue) == normalizeSectorName(scrapeValue)
	result.Details = fmt.Sprintf("API=%s, scrape=%s", apiValue, scrapeValue)
	return result
}

func probeCurrency(ticker string, dataset probeDataset) *ProbeResult {
	result := &ProbeResult{Ticker: ticker, ProbeType: "currency"}
	apiValue, apiOK := stringInfo(dataset.info, "financialCurrency", "currency")
	scrapeValue := ""
	if dataset.financials != nil {
		scrapeValue = dataset.financials.Currency
	}
	result.APIValue = apiValue
	result.ScrapeValue = scrapeValue
	if !apiOK || strings.TrimSpace(scrapeValue) == "" {
		result.Details = "currency is unavailable from one or both real sources"
		return result
	}
	result.Passed = strings.EqualFold(apiValue, scrapeValue)
	result.Details = fmt.Sprintf("API=%s, scrape=%s", apiValue, scrapeValue)
	return result
}

func comparePositive(
	result *ProbeResult,
	apiValue float64,
	apiOK bool,
	scrapeValue float64,
	scrapeOK bool,
) *ProbeResult {
	result.APIValue = apiValue
	result.ScrapeValue = scrapeValue
	if !apiOK || !scrapeOK || apiValue <= 0 || scrapeValue <= 0 {
		result.Details = "value is unavailable or non-positive from one or both real sources"
		return result
	}

	result.Difference = math.Abs(apiValue-scrapeValue) / apiValue
	result.Passed = result.Difference <= result.Tolerance
	result.Details = fmt.Sprintf(
		"API=%.4f, scrape=%.4f, difference=%.2f%%, tolerance=%.2f%%",
		apiValue,
		scrapeValue,
		result.Difference*100,
		result.Tolerance*100,
	)
	return result
}

func numericInfo(info map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		switch value := info[key].(type) {
		case float64:
			return value, true
		case float32:
			return float64(value), true
		case int:
			return float64(value), true
		case int64:
			return float64(value), true
		}
	}
	return 0, false
}

func stringInfo(info map[string]any, keys ...string) (string, bool) {
	for _, key := range keys {
		value, ok := info[key].(string)
		if ok && strings.TrimSpace(value) != "" {
			return value, true
		}
	}
	return "", false
}

func scaledValue(value *model.ScaledDecimal) (float64, bool) {
	if value == nil {
		return 0, false
	}
	return value.Float64(), true
}

func normalizeSectorName(sector string) string {
	sector = strings.ToLower(strings.TrimSpace(sector))
	replacements := map[string]string{
		"information technology": "technology",
		"tech":                   "technology",
		"healthcare":             "health care",
		"consumer discretionary": "consumer cyclical",
		"consumer staples":       "consumer defensive",
	}
	if normalized, exists := replacements[sector]; exists {
		return normalized
	}
	return sector
}
