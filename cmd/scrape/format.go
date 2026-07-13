// format.go — pure DTO → stdout formatters for the scrape subcommand's
// preview modes (analysis / analyst-insights). The `comprehensive-*` DTO
// formatters (stats / profile / financials) live in `format_comprehensive.go`
// to keep this file under the 700-line ceiling. Nothing here mutates state or
// talks to the network; it only renders a parsed DTO to human-readable lines.
//
// Capacity: printAnalysisSummary + printAnalysisRow + printAnalysisCell,
// printAnalystInsightsSummary.
package scrape

import (
	"fmt"

	"github.com/bizshuk/yfin/model"
)

// printAnalysisSummary prints a comprehensive summary of analysis data
func printAnalysisSummary(dto *model.ComprehensiveAnalysisDTO) {
	fmt.Printf("ANALYSIS SUMMARY: symbol=%s\n", dto.Symbol)

	// Earnings Estimate
	fmt.Printf("\nEARNINGS ESTIMATE (Currency: %s):\n", dto.EarningsEstimate.Currency)
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("No. of Analysts      ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.NoOfAnalysts, dto.EarningsEstimate.NextQtr.NoOfAnalysts,
		dto.EarningsEstimate.CurrentYear.NoOfAnalysts, dto.EarningsEstimate.NextYear.NoOfAnalysts, "int")
	fmt.Printf("Avg. Estimate        ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.AvgEstimate, dto.EarningsEstimate.NextQtr.AvgEstimate,
		dto.EarningsEstimate.CurrentYear.AvgEstimate, dto.EarningsEstimate.NextYear.AvgEstimate, "float")
	fmt.Printf("Low Estimate         ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.LowEstimate, dto.EarningsEstimate.NextQtr.LowEstimate,
		dto.EarningsEstimate.CurrentYear.LowEstimate, dto.EarningsEstimate.NextYear.LowEstimate, "float")
	fmt.Printf("High Estimate        ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.HighEstimate, dto.EarningsEstimate.NextQtr.HighEstimate,
		dto.EarningsEstimate.CurrentYear.HighEstimate, dto.EarningsEstimate.NextYear.HighEstimate, "float")
	fmt.Printf("Year Ago EPS         ")
	printAnalysisRow(dto.EarningsEstimate.CurrentQtr.YearAgoEPS, dto.EarningsEstimate.NextQtr.YearAgoEPS,
		dto.EarningsEstimate.CurrentYear.YearAgoEPS, dto.EarningsEstimate.NextYear.YearAgoEPS, "float")

	// Revenue Estimate
	fmt.Printf("\nREVENUE ESTIMATE (Currency: %s):\n", dto.RevenueEstimate.Currency)
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("No. of Analysts      ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.NoOfAnalysts, dto.RevenueEstimate.NextQtr.NoOfAnalysts,
		dto.RevenueEstimate.CurrentYear.NoOfAnalysts, dto.RevenueEstimate.NextYear.NoOfAnalysts, "int")
	fmt.Printf("Avg. Estimate        ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.AvgEstimate, dto.RevenueEstimate.NextQtr.AvgEstimate,
		dto.RevenueEstimate.CurrentYear.AvgEstimate, dto.RevenueEstimate.NextYear.AvgEstimate, "string")
	fmt.Printf("Low Estimate         ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.LowEstimate, dto.RevenueEstimate.NextQtr.LowEstimate,
		dto.RevenueEstimate.CurrentYear.LowEstimate, dto.RevenueEstimate.NextYear.LowEstimate, "string")
	fmt.Printf("High Estimate        ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.HighEstimate, dto.RevenueEstimate.NextQtr.HighEstimate,
		dto.RevenueEstimate.CurrentYear.HighEstimate, dto.RevenueEstimate.NextYear.HighEstimate, "string")
	fmt.Printf("Year Ago Sales       ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.YearAgoSales, dto.RevenueEstimate.NextQtr.YearAgoSales,
		dto.RevenueEstimate.CurrentYear.YearAgoSales, dto.RevenueEstimate.NextYear.YearAgoSales, "string")
	fmt.Printf("Sales Growth         ")
	printAnalysisRow(dto.RevenueEstimate.CurrentQtr.SalesGrowthYearEst, dto.RevenueEstimate.NextQtr.SalesGrowthYearEst,
		dto.RevenueEstimate.CurrentYear.SalesGrowthYearEst, dto.RevenueEstimate.NextYear.SalesGrowthYearEst, "string")

	// Earnings History
	fmt.Printf("\nEARNINGS HISTORY (Currency: %s):\n", dto.EarningsHistory.Currency)
	if len(dto.EarningsHistory.Data) > 0 {
		fmt.Printf("Date              EPS Est.    EPS Actual    Difference    Surprise %%\n")
		for _, entry := range dto.EarningsHistory.Data {
			fmt.Printf("%-16s  ", entry.Date)
			if entry.EPSEst != nil {
				fmt.Printf("%-10.2f  ", *entry.EPSEst)
			} else {
				fmt.Printf("%-10s  ", "--")
			}
			if entry.EPSActual != nil {
				fmt.Printf("%-10.2f  ", *entry.EPSActual)
			} else {
				fmt.Printf("%-10s  ", "--")
			}
			if entry.Difference != nil {
				fmt.Printf("%-10.2f  ", *entry.Difference)
			} else {
				fmt.Printf("%-10s  ", "--")
			}
			if entry.SurprisePercent != nil {
				fmt.Printf("%-10s", *entry.SurprisePercent)
			} else {
				fmt.Printf("%-10s", "--")
			}
			fmt.Printf("\n")
		}
	}

	// EPS Trend
	fmt.Printf("\nEPS TREND (Currency: %s):\n", dto.EPSTrend.Currency)
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("Current Estimate     ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.CurrentEstimate, dto.EPSTrend.NextQtr.CurrentEstimate,
		dto.EPSTrend.CurrentYear.CurrentEstimate, dto.EPSTrend.NextYear.CurrentEstimate, "float")
	fmt.Printf("7 Days Ago          ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.Days7Ago, dto.EPSTrend.NextQtr.Days7Ago,
		dto.EPSTrend.CurrentYear.Days7Ago, dto.EPSTrend.NextYear.Days7Ago, "float")
	fmt.Printf("30 Days Ago         ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.Days30Ago, dto.EPSTrend.NextQtr.Days30Ago,
		dto.EPSTrend.CurrentYear.Days30Ago, dto.EPSTrend.NextYear.Days30Ago, "float")
	fmt.Printf("60 Days Ago         ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.Days60Ago, dto.EPSTrend.NextQtr.Days60Ago,
		dto.EPSTrend.CurrentYear.Days60Ago, dto.EPSTrend.NextYear.Days60Ago, "float")
	fmt.Printf("90 Days Ago         ")
	printAnalysisRow(dto.EPSTrend.CurrentQtr.Days90Ago, dto.EPSTrend.NextQtr.Days90Ago,
		dto.EPSTrend.CurrentYear.Days90Ago, dto.EPSTrend.NextYear.Days90Ago, "float")

	// EPS Revisions
	fmt.Printf("\nEPS REVISIONS (Currency: %s):\n", dto.EPSRevisions.Currency)
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("Up Last 7 Days      ")
	printAnalysisRow(dto.EPSRevisions.CurrentQtr.UpLast7Days, dto.EPSRevisions.NextQtr.UpLast7Days,
		dto.EPSRevisions.CurrentYear.UpLast7Days, dto.EPSRevisions.NextYear.UpLast7Days, "int")
	fmt.Printf("Up Last 30 Days     ")
	printAnalysisRow(dto.EPSRevisions.CurrentQtr.UpLast30Days, dto.EPSRevisions.NextQtr.UpLast30Days,
		dto.EPSRevisions.CurrentYear.UpLast30Days, dto.EPSRevisions.NextYear.UpLast30Days, "int")
	fmt.Printf("Down Last 7 Days    ")
	printAnalysisRow(dto.EPSRevisions.CurrentQtr.DownLast7Days, dto.EPSRevisions.NextQtr.DownLast7Days,
		dto.EPSRevisions.CurrentYear.DownLast7Days, dto.EPSRevisions.NextYear.DownLast7Days, "int")
	fmt.Printf("Down Last 30 Days   ")
	printAnalysisRow(dto.EPSRevisions.CurrentQtr.DownLast30Days, dto.EPSRevisions.NextQtr.DownLast30Days,
		dto.EPSRevisions.CurrentYear.DownLast30Days, dto.EPSRevisions.NextYear.DownLast30Days, "int")

	// Growth Estimate
	fmt.Printf("\nGROWTH ESTIMATE:\n")
	fmt.Printf("                     Current Qtr    Next Qtr    Current Year    Next Year\n")
	fmt.Printf("Growth Rate          ")
	printAnalysisRow(dto.GrowthEstimate.CurrentQtr, dto.GrowthEstimate.NextQtr,
		dto.GrowthEstimate.CurrentYear, dto.GrowthEstimate.NextYear, "string")
}

// printAnalysisRow prints a formatted row for analysis tables
func printAnalysisRow(currentQtr, nextQtr, currentYear, nextYear interface{}, dataType string) {
	switch dataType {
	case "int":
		printAnalysisCell(currentQtr, "int")
		printAnalysisCell(nextQtr, "int")
		printAnalysisCell(currentYear, "int")
		printAnalysisCell(nextYear, "int")
	case "float":
		printAnalysisCell(currentQtr, "float")
		printAnalysisCell(nextQtr, "float")
		printAnalysisCell(currentYear, "float")
		printAnalysisCell(nextYear, "float")
	case "string":
		printAnalysisCell(currentQtr, "string")
		printAnalysisCell(nextQtr, "string")
		printAnalysisCell(currentYear, "string")
		printAnalysisCell(nextYear, "string")
	}
	fmt.Printf("\n")
}

// printAnalysisCell prints a single cell value with proper formatting
func printAnalysisCell(value interface{}, dataType string) {
	switch dataType {
	case "int":
		if v, ok := value.(*int); ok && v != nil {
			fmt.Printf("%-15d", *v)
		} else {
			fmt.Printf("%-15s", "--")
		}
	case "float":
		if v, ok := value.(*float64); ok && v != nil {
			fmt.Printf("%-15.2f", *v)
		} else {
			fmt.Printf("%-15s", "--")
		}
	case "string":
		if v, ok := value.(*string); ok && v != nil {
			fmt.Printf("%-15s", *v)
		} else {
			fmt.Printf("%-15s", "--")
		}
	}
}

// printAnalystInsightsSummary prints a comprehensive summary of analyst insights
func printAnalystInsightsSummary(dto *model.AnalystInsightsDTO) {
	fmt.Printf("ANALYST INSIGHTS: symbol=%s\n", dto.Symbol)

	// Current Price
	if dto.CurrentPrice != nil {
		fmt.Printf("Current Price: %.2f\n", *dto.CurrentPrice)
	}

	// Price Targets
	fmt.Printf("\nPRICE TARGETS:\n")
	if dto.TargetMeanPrice != nil {
		fmt.Printf("  Average Target: %.2f\n", *dto.TargetMeanPrice)
	}
	if dto.TargetMedianPrice != nil {
		fmt.Printf("  Median Target: %.2f\n", *dto.TargetMedianPrice)
	}
	if dto.TargetHighPrice != nil {
		fmt.Printf("  High Target: %.2f\n", *dto.TargetHighPrice)
	}
	if dto.TargetLowPrice != nil {
		fmt.Printf("  Low Target: %.2f\n", *dto.TargetLowPrice)
	}

	// Analyst Recommendations
	fmt.Printf("\nANALYST RECOMMENDATIONS:\n")
	if dto.NumberOfAnalysts != nil {
		fmt.Printf("  Number of Analysts: %d\n", *dto.NumberOfAnalysts)
	}
	if dto.RecommendationMean != nil {
		fmt.Printf("  Recommendation Score: %.2f\n", *dto.RecommendationMean)
	}
	if dto.RecommendationKey != nil {
		fmt.Printf("  Recommendation: %s\n", *dto.RecommendationKey)
	}

	// Calculate upside/downside potential
	if dto.CurrentPrice != nil && dto.TargetMeanPrice != nil {
		upside := ((*dto.TargetMeanPrice - *dto.CurrentPrice) / *dto.CurrentPrice) * 100
		fmt.Printf("\nPOTENTIAL:\n")
		if upside > 0 {
			fmt.Printf("  Upside Potential: +%.1f%%\n", upside)
		} else {
			fmt.Printf("  Downside Risk: %.1f%%\n", upside)
		}
	}
}
