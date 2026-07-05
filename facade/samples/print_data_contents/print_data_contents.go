// `print_data_contents.go` — field-by-field walk of `facade.Quote`,
// `facade.BarBatch`, news, financials, and key-statistics snapshots, including
// JSON marshaling. Capacity: ~167 LOC; APIs include `FetchQuote`,
// `FetchDailyBars`, `ScrapeNews`, `ScrapeFinancials`, `ScrapeKeyStatistics`.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bizshuk/yfin/facade"
)

// print_data_contents demonstrates the plain-struct SDK surface. As of Step 6
// of plans/spicy-singing-swan.md, facade.Client returns facade.Quote /
// facade.BarBatch / facade.FundamentalsSnapshot / facade.NewsItem — no
// reflection, no ScaledDecimal pointers. Each section below prints the public
// fields directly.
func main() {
	client := facade.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("contents-example-%d", time.Now().Unix())

	fmt.Println("=== SDK DATA CONTENTS (plain structs) ===")
	fmt.Println("Showing the contents of returned facade.* values")
	fmt.Println()

	// Example 1: Quote data contents
	fmt.Println("1. QUOTE DATA CONTENTS")
	fmt.Println("======================")
	quote, err := client.FetchQuote(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching quote: %v", err)
	} else {
		fmt.Printf("Quote for %s:\n", quote.Symbol)
		fmt.Printf("  Currency: %s\n", quote.Currency)
		fmt.Printf("  Price: %.4f\n", quote.Price)
		fmt.Printf("  Event Time: %s\n", quote.EventTime.Format(time.RFC3339))
	}

	// Example 2: Historical bars data contents
	fmt.Println("\n2. HISTORICAL BARS DATA CONTENTS")
	fmt.Println("================================")
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	dailyBars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
	if err != nil {
		log.Printf("Error fetching daily bars: %v", err)
	} else {
		fmt.Printf("Daily Bars for %s (%d bars):\n", dailyBars.Symbol, len(dailyBars.Bars))
		fmt.Printf("  MIC: %s\n", dailyBars.MIC)

		for i, bar := range dailyBars.Bars {
			if i >= 3 {
				fmt.Printf("  ... and %d more bars\n", len(dailyBars.Bars)-3)
				break
			}
			fmt.Printf("  Bar %d:\n", i+1)
			fmt.Printf("    Date: %s\n", bar.Date)
			fmt.Printf("    Open: %.4f\n", bar.Open)
			fmt.Printf("    High: %.4f\n", bar.High)
			fmt.Printf("    Low: %.4f\n", bar.Low)
			fmt.Printf("    Close: %.4f\n", bar.Close)
			fmt.Printf("    Volume: %d\n", bar.Volume)
			fmt.Printf("    Adjusted: %t\n", bar.Adjusted)
			fmt.Printf("    Currency: %s\n", bar.CurrencyCode)
		}
	}

	// Example 3: News data contents
	fmt.Println("\n3. NEWS DATA CONTENTS")
	fmt.Println("=====================")
	news, err := client.ScrapeNews(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching news: %v", err)
	} else {
		fmt.Printf("News for AAPL (%d articles):\n", len(news))

		for i, article := range news {
			if i >= 3 {
				fmt.Printf("  ... and %d more articles\n", len(news)-3)
				break
			}
			fmt.Printf("  Article %d:\n", i+1)
			fmt.Printf("    Headline: %s\n", article.Title)
			fmt.Printf("    Source: %s\n", article.Source)
			fmt.Printf("    Published: %s\n", article.PublishedAt.Format(time.RFC3339))
			if len(article.Symbols) > 0 {
				fmt.Printf("    Related Symbols: %v\n", article.Symbols)
			}
			fmt.Printf("    URL: %s\n", article.URL)
		}
	}

	// Example 4: Financials data contents
	fmt.Println("\n4. FINANCIALS DATA CONTENTS")
	fmt.Println("===========================")
	financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching financials: %v", err)
	} else {
		fmt.Printf("Financials for %s (%d lines):\n", financials.Symbol, len(financials.Lines))
		fmt.Printf("  MIC: %s\n", financials.MIC)
		fmt.Printf("  Source: %s\n", financials.Source)
		fmt.Printf("  As Of: %s\n", financials.AsOf.Format(time.RFC3339))

		for i, line := range financials.Lines {
			if i >= 5 {
				fmt.Printf("  ... and %d more lines\n", len(financials.Lines)-5)
				break
			}
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    Key: %s\n", line.Key)
			fmt.Printf("    Value: %.2f\n", line.Value)
			fmt.Printf("    Currency: %s\n", line.CurrencyCode)
			fmt.Printf("    Period: %s to %s\n",
				line.PeriodStart.Format("2006-01-02"),
				line.PeriodEnd.Format("2006-01-02"))
		}
	}

	// Example 5: Key Statistics data contents
	fmt.Println("\n5. KEY STATISTICS DATA CONTENTS")
	fmt.Println("===============================")
	keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching key statistics: %v", err)
	} else {
		fmt.Printf("Key Statistics for %s (%d lines):\n", keyStats.Symbol, len(keyStats.Lines))
		fmt.Printf("  Source: %s\n", keyStats.Source)
		fmt.Printf("  As Of: %s\n", keyStats.AsOf.Format(time.RFC3339))

		for i, line := range keyStats.Lines {
			if i >= 5 {
				fmt.Printf("  ... and %d more lines\n", len(keyStats.Lines)-5)
				break
			}
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    Key: %s\n", line.Key)
			fmt.Printf("    Value: %.2f\n", line.Value)
			fmt.Printf("    Currency: %s\n", line.CurrencyCode)
			fmt.Printf("    Period: %s to %s\n",
				line.PeriodStart.Format("2006-01-02"),
				line.PeriodEnd.Format("2006-01-02"))
		}
	}

	// Example 6: Show JSON representation of one message
	fmt.Println("\n6. JSON REPRESENTATION EXAMPLE")
	fmt.Println("==============================")
	if quote != nil {
		fmt.Println("Quote data as JSON:")
		jsonData, err := json.MarshalIndent(quote, "", "  ")
		if err != nil {
			log.Printf("Error marshaling to JSON: %v", err)
		} else {
			fmt.Println(string(jsonData))
		}
	}

	fmt.Println("\n=== DATA CONTENTS COMPLETE ===")
	fmt.Println("All data is plain Go structs with:")
	fmt.Println("- float64 prices (decoded from internal ScaledDecimal)")
	fmt.Println("- UTC timestamps")
	fmt.Println("- nil-aware nullable fields where appropriate")
}
