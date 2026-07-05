package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bizshuk/yfin/facade"
)

// api_usage demonstrates the plain-struct SDK surface. As of Step 6 of
// plans/spicy-singing-swan.md, every Scrape*/Fetch* call returns a plain
// facade.* struct instead of an ampy-proto message; this file prints
// only the public fields (Symbol, Source, Lines[i].Key/Value/Period, etc.).
func main() {
	client := facade.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("api-example-%d", time.Now().Unix())

	fmt.Println("=== SDK API USAGE (plain structs) ===")
	fmt.Println("Demonstrating programmatic access via the SDK facade")
	fmt.Println()

	// Example 1: Fetch financials data
	fmt.Println("1. Fetching Financials Data...")
	financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching financials: %v", err)
	} else {
		fmt.Printf("Financials: %d lines, source: %s, as_of: %s\n",
			len(financials.Lines), financials.Source, financials.AsOf.Format(time.RFC3339))
	}

	// Example 2: Fetch key statistics
	fmt.Println("\n2. Fetching Key Statistics...")
	keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching key statistics: %v", err)
	} else {
		fmt.Printf("Key Statistics: %d lines, source: %s, as_of: %s\n",
			len(keyStats.Lines), keyStats.Source, keyStats.AsOf.Format(time.RFC3339))
	}

	// Example 3: Fetch analysis data
	fmt.Println("\n3. Fetching Analysis Data...")
	analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching analysis: %v", err)
	} else {
		fmt.Printf("Analysis: %d lines, source: %s, as_of: %s\n",
			len(analysis.Lines), analysis.Source, analysis.AsOf.Format(time.RFC3339))
	}

	// Example 4: Fetch news data
	fmt.Println("\n4. Fetching News Data...")
	news, err := client.ScrapeNews(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching news: %v", err)
	} else {
		fmt.Printf("News: %d articles\n", len(news))
	}

	// Example 5: Fetch all fundamentals at once
	fmt.Println("\n5. Fetching All Fundamentals...")
	allFundamentals, err := client.ScrapeAllFundamentals(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching all fundamentals: %v", err)
	} else {
		fmt.Printf("All Fundamentals: %d snapshots\n", len(allFundamentals))
		for i, snapshot := range allFundamentals {
			fmt.Printf("   %d. source=%s, lines=%d\n",
				i+1, snapshot.Source, len(snapshot.Lines))
		}
	}

	fmt.Println("\n=== API EXAMPLE COMPLETE ===")
	fmt.Println("All data is returned as plain Go structs (facade.BarBatch,")
	fmt.Println("facade.Quote, facade.FundamentalsSnapshot, facade.NewsItem, etc.).")
	fmt.Println("They marshal straight to JSON, no reflection on internal types.")
}
