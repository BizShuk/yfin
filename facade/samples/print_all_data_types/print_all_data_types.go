package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/bizshuk/yfin/facade"
)

// print_all_data_types demonstrates the plain-struct SDK surface. As of Step 6
// of plans/spicy-singing-swan.md, facade.Client.Scrape* returns
// facade.FundamentalsSnapshot (or facade.FundamentalsSnapshot[] from
// ScrapeAllFundamentals) instead of ampy-proto messages. Lines are plain
// facade.FundamentalsLine values (float64 + time.Time, no ScaledDecimal or
// Timestamp protobuf fields).
func main() {
	client := facade.NewClient()
	ctx := context.Background()
	runID := fmt.Sprintf("all-types-example-%d", time.Now().Unix())

	fmt.Println("=== ALL SDK DATA TYPES CONTENTS (plain structs) ===")
	fmt.Println("Showing contents of all available facade.* struct types")
	fmt.Println()

	// Example 1: Analysis data contents
	fmt.Println("1. ANALYSIS DATA CONTENTS")
	fmt.Println("=========================")
	analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching analysis: %v", err)
	} else {
		fmt.Printf("Analysis for %s (%d lines):\n", analysis.Symbol, len(analysis.Lines))
		fmt.Printf("  Source: %s\n", analysis.Source)
		fmt.Printf("  As Of: %s\n", analysis.AsOf.Format(time.RFC3339))

		for i, line := range analysis.Lines {
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    Key: %s\n", line.Key)
			fmt.Printf("    Value: %.2f\n", line.Value)
			fmt.Printf("    Currency: %s\n", line.CurrencyCode)
			fmt.Printf("    Period: %s to %s\n",
				line.PeriodStart.Format("2006-01-02"),
				line.PeriodEnd.Format("2006-01-02"))
		}
	}

	// Example 2: Analyst Insights data contents
	fmt.Println("\n2. ANALYST INSIGHTS DATA CONTENTS")
	fmt.Println("==================================")
	analystInsights, err := client.ScrapeAnalystInsights(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching analyst insights: %v", err)
	} else {
		fmt.Printf("Analyst Insights for %s (%d lines):\n", analystInsights.Symbol, len(analystInsights.Lines))
		fmt.Printf("  Source: %s\n", analystInsights.Source)
		fmt.Printf("  As Of: %s\n", analystInsights.AsOf.Format(time.RFC3339))

		for i, line := range analystInsights.Lines {
			fmt.Printf("  Line %d:\n", i+1)
			fmt.Printf("    Key: %s\n", line.Key)
			fmt.Printf("    Value: %.2f\n", line.Value)
			fmt.Printf("    Currency: %s\n", line.CurrencyCode)
			fmt.Printf("    Period: %s to %s\n",
				line.PeriodStart.Format("2006-01-02"),
				line.PeriodEnd.Format("2006-01-02"))
		}
	}

	// Example 3: Balance Sheet data contents
	fmt.Println("\n3. BALANCE SHEET DATA CONTENTS")
	fmt.Println("==============================")
	balanceSheet, err := client.ScrapeBalanceSheet(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching balance sheet: %v", err)
	} else {
		fmt.Printf("Balance Sheet for %s (%d lines):\n", balanceSheet.Symbol, len(balanceSheet.Lines))
		fmt.Printf("  Source: %s\n", balanceSheet.Source)
		fmt.Printf("  As Of: %s\n", balanceSheet.AsOf.Format(time.RFC3339))

		for i, line := range balanceSheet.Lines {
			if i >= 5 {
				fmt.Printf("  ... and %d more lines\n", len(balanceSheet.Lines)-5)
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

	// Example 4: Cash Flow data contents
	fmt.Println("\n4. CASH FLOW DATA CONTENTS")
	fmt.Println("===========================")
	cashFlow, err := client.ScrapeCashFlow(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching cash flow: %v", err)
	} else {
		fmt.Printf("Cash Flow for %s (%d lines):\n", cashFlow.Symbol, len(cashFlow.Lines))
		fmt.Printf("  Source: %s\n", cashFlow.Source)
		fmt.Printf("  As Of: %s\n", cashFlow.AsOf.Format(time.RFC3339))

		for i, line := range cashFlow.Lines {
			if i >= 5 {
				fmt.Printf("  ... and %d more lines\n", len(cashFlow.Lines)-5)
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

	// Example 5: Show JSON representation of fundamentals data
	fmt.Println("\n5. JSON REPRESENTATION EXAMPLE (ANALYSIS)")
	fmt.Println("==========================================")
	if analysis != nil {
		fmt.Println("Analysis data as JSON:")
		jsonData, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			log.Printf("Error marshaling to JSON: %v", err)
		} else {
			fmt.Println(string(jsonData))
		}
	}

	// Example 6: Show all fundamentals at once with source differentiation
	fmt.Println("\n6. ALL FUNDAMENTALS WITH SOURCE DIFFERENTIATION")
	fmt.Println("==============================================")
	allFundamentals, err := client.ScrapeAllFundamentals(ctx, "AAPL", runID)
	if err != nil {
		log.Printf("Error fetching all fundamentals: %v", err)
	} else {
		fmt.Printf("All Fundamentals for AAPL (%d snapshots):\n", len(allFundamentals))
		for i, snapshot := range allFundamentals {
			fmt.Printf("  %d. source=%s, lines=%d, as_of=%s\n",
				i+1, snapshot.Source, len(snapshot.Lines),
				snapshot.AsOf.Format(time.RFC3339))

			if len(snapshot.Lines) > 0 {
				firstLine := snapshot.Lines[0]
				fmt.Printf("     Example Line: %s = %.2f %s\n",
					firstLine.Key, firstLine.Value, firstLine.CurrencyCode)
			}
			fmt.Println()
		}
	}

	fmt.Println("=== ALL DATA TYPES COMPLETE ===")
	fmt.Println("Summary of plain SDK struct types exposed by facade.Client:")
	fmt.Println("1. Quotes → facade.Quote{Symbol, Price, Currency, EventTime}")
	fmt.Println("2. Historical Bars → facade.BarBatch{Symbol, MIC, Bars []Bar}")
	fmt.Println("3. News → []facade.NewsItem{Title, URL, Source, PublishedAt, Symbols}")
	fmt.Println("4. Financials → facade.FundamentalsSnapshot")
	fmt.Println("5. Balance Sheet → facade.FundamentalsSnapshot (Source: yfinance/scrape/balance-sheet)")
	fmt.Println("6. Cash Flow → facade.FundamentalsSnapshot (Source: yfinance/scrape/cash-flow)")
	fmt.Println("7. Key Statistics → facade.FundamentalsSnapshot")
	fmt.Println("8. Analysis → facade.FundamentalsSnapshot")
	fmt.Println("9. Analyst Insights → facade.FundamentalsSnapshot")
	fmt.Println()
	fmt.Println("Source differentiation via the FundamentalsSnapshot.Source field:")
	fmt.Println("- yfinance/scrape/financials")
	fmt.Println("- yfinance/scrape/balance-sheet")
	fmt.Println("- yfinance/scrape/cash-flow")
	fmt.Println("- yfinance/scrape/key-statistics")
	fmt.Println("- yfinance/scrape/analysis")
	fmt.Println("- yfinance/scrape/analyst-insights")
}
