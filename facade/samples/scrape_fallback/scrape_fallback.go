package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bizshuk/yfinance-go/facade"
	"github.com/bizshuk/yfinance-go/utils/httpx"
)

// scrape_fallback demonstrates the plain-struct SDK surface. As of Step 6 of
// plans/spicy-singing-swan.md, facade.Client.Scrape* returns
// facade.FundamentalsSnapshot / []facade.NewsItem — the field accesses below
// use Symbol / Lines[i].Key / Source / AsOf instead of the previous
// Security.Symbol / Lines[i].Value (proto Decimal) / Meta.SchemaVersion.

// Example 1: Basic Scrape Fallback Usage — the simplest way to use the scrape
// fallback system.
func basicScrapeExample() {
	fmt.Println("=== Example 1: Basic Scrape Fallback ===")

	client := facade.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("example-basic-%d", time.Now().Unix())

	keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error scraping key statistics: %v", err)
		return
	}

	fmt.Printf("Successfully scraped key statistics for AAPL\n")
	fmt.Printf("Source: %s\n", keyStats.Source)
	fmt.Printf("Run ID echo (since facade dropped Meta.RunID): %s\n", runID)

	if len(keyStats.Lines) > 0 {
		fmt.Printf("Found %d data lines:\n", len(keyStats.Lines))
		for i, line := range keyStats.Lines {
			if i >= 3 {
				fmt.Printf("... and %d more lines\n", len(keyStats.Lines)-3)
				break
			}
			fmt.Printf("  - %s: %.4f\n", line.Key, line.Value)
		}
	}
}

// Example 2: Comprehensive Data Collection — multiple data types with error
// handling.
func comprehensiveDataExample() {
	fmt.Println("\n=== Example 2: Comprehensive Data Collection ===")

	client := facade.NewClient()
	ctx := context.Background()
	ticker := "MSFT"
	runID := fmt.Sprintf("example-comprehensive-%d", time.Now().Unix())

	dataTypes := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Key Statistics",
			fn: func() error {
				data, err := client.ScrapeKeyStatistics(ctx, ticker, runID)
				if err != nil {
					return err
				}
				fmt.Printf("  ✓ Key Statistics: %d lines collected\n", len(data.Lines))
				return nil
			},
		},
		{
			name: "Financials",
			fn: func() error {
				data, err := client.ScrapeFinancials(ctx, ticker, runID)
				if err != nil {
					return err
				}
				fmt.Printf("  ✓ Financials: %d lines collected\n", len(data.Lines))
				return nil
			},
		},
		{
			name: "Analysis",
			fn: func() error {
				data, err := client.ScrapeAnalysis(ctx, ticker, runID)
				if err != nil {
					return err
				}
				fmt.Printf("  ✓ Analysis: %d lines collected\n", len(data.Lines))
				return nil
			},
		},
		{
			name: "News",
			fn: func() error {
				articles, err := client.ScrapeNews(ctx, ticker, runID)
				if err != nil {
					return err
				}
				fmt.Printf("  ✓ News: %d articles collected\n", len(articles))
				return nil
			},
		},
	}

	fmt.Printf("Collecting data for %s...\n", ticker)
	successCount := 0

	for _, dataType := range dataTypes {
		if err := dataType.fn(); err != nil {
			fmt.Printf("  ✗ %s: %v\n", dataType.name, err)
		} else {
			successCount++
		}

		// Rate limiting - wait between requests
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("Completed: %d/%d data types collected successfully\n", successCount, len(dataTypes))
}

// Example 3: Advanced Configuration and Error Handling — robust error
// handling across multiple markets.
func advancedConfigExample() {
	fmt.Println("\n=== Example 3: Advanced Configuration ===")

	config := &httpx.Config{
		Timeout:     45 * time.Second,
		MaxAttempts: 3,
		QPS:         1.0,
		UserAgent:   "MyApp/1.0 (contact@mycompany.com)",
	}

	client := facade.NewClientWithConfig(config)
	ctx := context.Background()
	runID := fmt.Sprintf("example-advanced-%d", time.Now().Unix())

	tickers := []struct {
		symbol      string
		description string
	}{
		{"AAPL", "US Large Cap"},
		{"TSLA", "US Growth Stock"},
		{"0700.HK", "International (Hong Kong)"},
		{"SAP", "European Stock"},
	}

	fmt.Println("Testing scraping across different market types...")

	for _, ticker := range tickers {
		fmt.Printf("\nTesting %s (%s):\n", ticker.symbol, ticker.description)

		startTime := time.Now()
		keyStats, err := client.ScrapeKeyStatistics(ctx, ticker.symbol, runID)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("  ✗ Failed after %v: %v\n", duration, err)

			switch {
			case isRateLimitError(err):
				fmt.Printf("    → Rate limit detected, should implement backoff\n")
			case isParseError(err):
				fmt.Printf("    → Parse error, may indicate schema drift\n")
			case isNetworkError(err):
				fmt.Printf("    → Network error, should retry\n")
			default:
				fmt.Printf("    → Unknown error type\n")
			}
			continue
		}

		fmt.Printf("  ✓ Success in %v\n", duration)
		fmt.Printf("    Lines: %d\n", len(keyStats.Lines))
		fmt.Printf("    Source: %s\n", keyStats.Source)

		time.Sleep(2 * time.Second)
	}
}

// Example 4: Batch Processing with Orchestration — process multiple tickers
// efficiently.
func batchProcessingExample() {
	fmt.Println("\n=== Example 4: Batch Processing ===")

	client := facade.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("example-batch-%d", time.Now().Unix())

	universe := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA"}

	type Result struct {
		Ticker   string
		Success  bool
		Duration time.Duration
		Error    error
		Metrics  int
	}

	results := make([]Result, 0, len(universe))

	fmt.Printf("Processing %d tickers in batch...\n", len(universe))

	for i, ticker := range universe {
		fmt.Printf("[%d/%d] Processing %s...", i+1, len(universe), ticker)

		startTime := time.Now()
		keyStats, err := client.ScrapeKeyStatistics(ctx, ticker, runID)
		duration := time.Since(startTime)

		result := Result{
			Ticker:   ticker,
			Duration: duration,
		}

		if err != nil {
			result.Success = false
			result.Error = err
			fmt.Printf(" ✗ (%.2fs)\n", duration.Seconds())
		} else {
			result.Success = true
			result.Metrics = len(keyStats.Lines)
			fmt.Printf(" ✓ (%.2fs, %d metrics)\n", duration.Seconds(), result.Metrics)
		}

		results = append(results, result)

		if i < len(universe)-1 {
			time.Sleep(1 * time.Second)
		}
	}

	fmt.Println("\n=== Batch Processing Summary ===")
	successCount := 0
	totalDuration := time.Duration(0)
	totalMetrics := 0

	for _, result := range results {
		totalDuration += result.Duration
		if result.Success {
			successCount++
			totalMetrics += result.Metrics
		}
	}

	fmt.Printf("Success Rate: %d/%d (%.1f%%)\n", successCount, len(results), float64(successCount)/float64(len(results))*100)
	fmt.Printf("Average Duration: %.2fs\n", totalDuration.Seconds()/float64(len(results)))
	fmt.Printf("Total Metrics Collected: %d\n", totalMetrics)

	if successCount < len(results) {
		fmt.Println("\nFailures:")
		for _, result := range results {
			if !result.Success {
				fmt.Printf("  %s: %v\n", result.Ticker, result.Error)
			}
		}
	}
}

// Example 5: Real-time Data Pipeline Integration — feed facade.* into a
// downstream pipeline.
func pipelineIntegrationExample() {
	fmt.Println("\n=== Example 5: Pipeline Integration ===")

	client := facade.NewClient()
	ctx := context.Background()

	type DataPipeline struct {
		client *facade.Client
		runID  string
	}

	pipeline := &DataPipeline{
		client: client,
		runID:  fmt.Sprintf("pipeline-%d", time.Now().Unix()),
	}

	processTicker := func(ticker string) error {
		fmt.Printf("Pipeline processing %s...\n", ticker)

		keyStats, err := pipeline.client.ScrapeKeyStatistics(ctx, ticker, pipeline.runID)
		if err != nil {
			return fmt.Errorf("key statistics failed: %w", err)
		}

		financials, err := pipeline.client.ScrapeFinancials(ctx, ticker, pipeline.runID)
		if err != nil {
			return fmt.Errorf("financials failed: %w", err)
		}

		var news []facade.NewsItem
		n, err := pipeline.client.ScrapeNews(ctx, ticker, pipeline.runID)
		if err != nil {
			fmt.Printf("  Warning: News collection failed: %v\n", err)
		} else {
			news = n
		}

		fmt.Printf("  ✓ Key Statistics: %d lines\n", len(keyStats.Lines))
		fmt.Printf("  ✓ Financials: %d lines\n", len(financials.Lines))
		fmt.Printf("  ✓ News: %d articles\n", len(news))

		if len(keyStats.Lines) == 0 {
			return fmt.Errorf("no key statistics found")
		}

		if len(financials.Lines) == 0 {
			return fmt.Errorf("no financial data found")
		}

		fmt.Printf("  ✓ Data validated and ready for storage\n")

		return nil
	}

	testTickers := []string{"AAPL", "MSFT", "GOOGL"}

	for _, ticker := range testTickers {
		if err := processTicker(ticker); err != nil {
			fmt.Printf("  ✗ Pipeline failed for %s: %v\n", ticker, err)
		} else {
			fmt.Printf("  ✓ Pipeline completed successfully for %s\n", ticker)
		}

		time.Sleep(2 * time.Second)
	}
}

// Helper functions for error classification
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "429") || contains(errStr, "rate limit") || contains(errStr, "too many requests")
}

func isParseError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "parse") || contains(errStr, "schema") || contains(errStr, "extract")
}

func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "timeout") || contains(errStr, "connection") || contains(errStr, "network")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				indexOfSubstring(s, substr) >= 0))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func main() {
	fmt.Println("yfinance-go Scrape Fallback Examples (plain structs)")
	fmt.Println("=====================================================")

	basicScrapeExample()
	comprehensiveDataExample()
	advancedConfigExample()
	batchProcessingExample()
	pipelineIntegrationExample()

	fmt.Println("\n=== All Examples Completed ===")
	fmt.Println("For more information, see:")
	fmt.Println("- Documentation: docs/scrape/")
	fmt.Println("- Configuration: docs/scrape/config.md")
	fmt.Println("- Troubleshooting: docs/scrape/troubleshooting.md")
}