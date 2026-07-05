package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bizshuk/yfin/facade"
)

// historical_data_example demonstrates the plain-struct SDK surface. As of
// Step 6 of plans/spicy-singing-swan.md, facade.Client returns reflection-free
// structs (facade.BarBatch, facade.Quote) instead of internal/norm types; this
// file prints the new public fields directly.
func main() {
	client := facade.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("historical-example-%d", time.Now().Unix())

	fmt.Println("=== HISTORICAL DATA API EXAMPLE ===")
	fmt.Println("Demonstrating programmatic access via the SDK facade")
	fmt.Println()

	// Example 1: Fetch daily bars
	fmt.Println("1. Fetching Daily Bars...")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	dailyBars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
	if err != nil {
		log.Printf("Error fetching daily bars: %v", err)
	} else {
		fmt.Printf("Daily Bars: %d bars\n", len(dailyBars.Bars))
		if len(dailyBars.Bars) > 0 {
			fmt.Printf("   first close=%.2f %s  last close=%.2f %s\n",
				dailyBars.Bars[0].Close, dailyBars.Bars[0].CurrencyCode,
				dailyBars.Bars[len(dailyBars.Bars)-1].Close,
				dailyBars.Bars[len(dailyBars.Bars)-1].CurrencyCode)
		}
	}

	// Example 2: Fetch quote
	fmt.Println("\n2. Fetching Quote...")
	quote, err := client.FetchQuote(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching quote: %v", err)
	} else {
		fmt.Printf("Quote: symbol=%s price=%.4f %s event_time=%s\n",
			quote.Symbol, quote.Price, quote.Currency,
			quote.EventTime.Format(time.RFC3339))
	}

	// Example 3: Fetch intraday bars
	fmt.Println("\n3. Fetching Intraday Bars (1m)...")
	intradayBars, err := client.FetchIntradayBars(ctx, "AAPL", start, end, "1m", runID)
	if err != nil {
		log.Printf("Error fetching intraday bars: %v", err)
	} else {
		fmt.Printf("Intraday Bars: %d bars\n", len(intradayBars.Bars))
	}

	fmt.Println("\n=== HISTORICAL DATA EXAMPLE COMPLETE ===")
	fmt.Println("All data is returned as plain Go structs:")
	fmt.Println("- Daily/Weekly/Monthly bars → facade.BarBatch{Bars []facade.Bar}")
	fmt.Println("- Quotes → facade.Quote{Symbol, Price, Currency, EventTime}")
	fmt.Println("- Intraday bars → facade.BarBatch")
	fmt.Println("- Company info → facade.CompanyInfo")
	fmt.Println("- Market data → facade.MarketData")
}
