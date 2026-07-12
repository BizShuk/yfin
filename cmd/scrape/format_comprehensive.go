// format_comprehensive.go — comprehensive-* DTO → stdout formatters used by
// the scrape preview-json case "key-statistics" / "profile" / "financials" /
// "balance-sheet" / "cash-flow". The same formatters exist in the
// fundamentals sub-package for the comprehensive-stats / comprehensive-profile
// subcommands; duplicated here to avoid a scrape → fundamentals cross-package
// dependency.
//
// Capacity: printComprehensiveStatisticsSummary,
// printComprehensiveProfileSummary, printComprehensiveFinancialsSummary,
// countFinancialsFields.
package scrape

import (
	"fmt"

	"github.com/bizshuk/yfin/svc/scrape"
)

// printComprehensiveStatisticsSummary prints a summary of comprehensive statistics
func printComprehensiveStatisticsSummary(dto *scrape.ComprehensiveKeyStatisticsDTO) {
	fmt.Printf("COMPREHENSIVE STATISTICS: symbol=%s currency=%s\n", dto.Symbol, dto.Currency)

	// Current values
	fmt.Printf("CURRENT VALUES:\n")
	if dto.Current.MarketCap != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.MarketCap.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.MarketCap.Scaled) / multiplier
		fmt.Printf("  Market Cap: %.2fB\n", actualValue/1e9)
	}
	if dto.Current.EnterpriseValue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValue.Scaled) / multiplier
		fmt.Printf("  Enterprise Value: %.2fB\n", actualValue/1e9)
	}
	if dto.Current.ForwardPE != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.ForwardPE.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.ForwardPE.Scaled) / multiplier
		fmt.Printf("  Forward P/E: %.2f\n", actualValue)
	}
	if dto.Current.TrailingPE != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TrailingPE.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TrailingPE.Scaled) / multiplier
		fmt.Printf("  Trailing P/E: %.2f\n", actualValue)
	}
	if dto.Current.PEGRatio != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PEGRatio.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PEGRatio.Scaled) / multiplier
		fmt.Printf("  PEG Ratio: %.2f\n", actualValue)
	}
	if dto.Current.PriceSales != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PriceSales.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PriceSales.Scaled) / multiplier
		fmt.Printf("  Price/Sales: %.2f\n", actualValue)
	}
	if dto.Current.PriceBook != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.PriceBook.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.PriceBook.Scaled) / multiplier
		fmt.Printf("  Price/Book: %.2f\n", actualValue)
	}
	if dto.Current.EnterpriseValueRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValueRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValueRevenue.Scaled) / multiplier
		fmt.Printf("  Enterprise Value/Revenue: %.2f\n", actualValue)
	}
	if dto.Current.EnterpriseValueEBITDA != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EnterpriseValueEBITDA.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EnterpriseValueEBITDA.Scaled) / multiplier
		fmt.Printf("  Enterprise Value/EBITDA: %.2f\n", actualValue)
	}

	// Additional statistics
	fmt.Printf("ADDITIONAL STATISTICS:\n")
	if dto.Additional.Beta != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.Beta.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.Beta.Scaled) / multiplier
		fmt.Printf("  Beta: %.2f\n", actualValue)
	}
	if dto.Additional.SharesOutstanding != nil {
		fmt.Printf("  Shares Outstanding: %.2fB\n", float64(*dto.Additional.SharesOutstanding)/1e9)
	}
	if dto.Additional.ProfitMargin != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ProfitMargin.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ProfitMargin.Scaled) / multiplier
		fmt.Printf("  Profit Margin: %.2f%%\n", actualValue)
	}
	if dto.Additional.OperatingMargin != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.OperatingMargin.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.OperatingMargin.Scaled) / multiplier
		fmt.Printf("  Operating Margin: %.2f%%\n", actualValue)
	}
	if dto.Additional.ReturnOnAssets != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ReturnOnAssets.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ReturnOnAssets.Scaled) / multiplier
		fmt.Printf("  Return on Assets: %.2f%%\n", actualValue)
	}
	if dto.Additional.ReturnOnEquity != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Additional.ReturnOnEquity.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Additional.ReturnOnEquity.Scaled) / multiplier
		fmt.Printf("  Return on Equity: %.2f%%\n", actualValue)
	}

	// Historical values
	if len(dto.Historical) > 0 {
		fmt.Printf("HISTORICAL VALUES:\n")
		for _, quarter := range dto.Historical {
			fmt.Printf("  %s:\n", quarter.Date)
			if quarter.MarketCap != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.MarketCap.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.MarketCap.Scaled) / multiplier
				fmt.Printf("    Market Cap: %.2fB\n", actualValue/1e9)
			}
			if quarter.ForwardPE != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.ForwardPE.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.ForwardPE.Scaled) / multiplier
				fmt.Printf("    Forward P/E: %.2f\n", actualValue)
			}
			if quarter.TrailingPE != nil {
				multiplier := float64(1)
				for i := 0; i < quarter.TrailingPE.Scale; i++ {
					multiplier *= 10
				}
				actualValue := float64(quarter.TrailingPE.Scaled) / multiplier
				fmt.Printf("    Trailing P/E: %.2f\n", actualValue)
			}
		}
	}
}

// printComprehensiveProfileSummary prints a summary of comprehensive profile
func printComprehensiveProfileSummary(dto *scrape.ComprehensiveProfileDTO) {
	fmt.Printf("COMPREHENSIVE PROFILE: symbol=%s\n", dto.Symbol)

	// Company Information
	fmt.Printf("COMPANY INFORMATION:\n")
	if dto.CompanyName != "" {
		fmt.Printf("  Company Name: %s\n", dto.CompanyName)
	}
	if dto.ShortName != "" {
		fmt.Printf("  Short Name: %s\n", dto.ShortName)
	}
	if dto.Address1 != "" {
		fmt.Printf("  Address: %s\n", dto.Address1)
	}
	if dto.City != "" && dto.State != "" {
		fmt.Printf("  City, State: %s, %s\n", dto.City, dto.State)
	}
	if dto.Zip != "" {
		fmt.Printf("  ZIP: %s\n", dto.Zip)
	}
	if dto.Country != "" {
		fmt.Printf("  Country: %s\n", dto.Country)
	}
	if dto.Phone != "" {
		fmt.Printf("  Phone: %s\n", dto.Phone)
	}
	if dto.Website != "" {
		fmt.Printf("  Website: %s\n", dto.Website)
	}
	if dto.Industry != "" {
		fmt.Printf("  Industry: %s\n", dto.Industry)
	}
	if dto.Sector != "" {
		fmt.Printf("  Sector: %s\n", dto.Sector)
	}
	if dto.FullTimeEmployees != nil {
		fmt.Printf("  Full Time Employees: %d\n", *dto.FullTimeEmployees)
	}
	if dto.BusinessSummary != "" {
		summary := dto.BusinessSummary
		if len(summary) > 200 {
			summary = summary[:200] + "..."
		}
		fmt.Printf("  Business Summary: %s\n", summary)
	}

	// Key Executives
	if len(dto.Executives) > 0 {
		fmt.Printf("KEY EXECUTIVES:\n")
		for i, exec := range dto.Executives {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s", i+1, exec.Name)
			if exec.Title != "" {
				fmt.Printf(" - %s", exec.Title)
			}
			if exec.YearBorn != nil {
				fmt.Printf(" (Born: %d)", *exec.YearBorn)
			}
			if exec.TotalPay != nil {
				fmt.Printf(" - Total Pay: $%.2fM", float64(*exec.TotalPay)/1e6)
			}
			fmt.Printf("\n")
		}
	}

	// Additional Information
	fmt.Printf("ADDITIONAL INFORMATION:\n")
	if dto.MaxAge != nil {
		fmt.Printf("  Max Age: %d\n", *dto.MaxAge)
	}
	if dto.AuditRisk != nil {
		fmt.Printf("  Audit Risk: %d\n", *dto.AuditRisk)
	}
	if dto.BoardRisk != nil {
		fmt.Printf("  Board Risk: %d\n", *dto.BoardRisk)
	}
	if dto.CompensationRisk != nil {
		fmt.Printf("  Compensation Risk: %d\n", *dto.CompensationRisk)
	}
	if dto.ShareHolderRightsRisk != nil {
		fmt.Printf("  Share Holder Rights Risk: %d\n", *dto.ShareHolderRightsRisk)
	}
	if dto.OverallRisk != nil {
		fmt.Printf("  Overall Risk: %d\n", *dto.OverallRisk)
	}
}

// printComprehensiveFinancialsSummary prints a summary of comprehensive financials
func printComprehensiveFinancialsSummary(dto *scrape.ComprehensiveFinancialsDTO) {
	fmt.Printf("COMPREHENSIVE FINANCIALS: symbol=%s currency=%s\n", dto.Symbol, dto.Currency)

	// Current values
	fmt.Printf("CURRENT VALUES:\n")
	if dto.Current.TotalRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalRevenue.Scaled) / multiplier
		fmt.Printf("  Total Revenue: %.0f\n", actualValue)
	}
	if dto.Current.CostOfRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.CostOfRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.CostOfRevenue.Scaled) / multiplier
		fmt.Printf("  Cost of Revenue: %.0f\n", actualValue)
	}
	if dto.Current.GrossProfit != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.GrossProfit.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.GrossProfit.Scaled) / multiplier
		fmt.Printf("  Gross Profit: %.0f\n", actualValue)
	}
	if dto.Current.OperatingIncome != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.OperatingIncome.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.OperatingIncome.Scaled) / multiplier
		fmt.Printf("  Operating Income: %.0f\n", actualValue)
	}
	if dto.Current.NetIncomeCommonStockholders != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.NetIncomeCommonStockholders.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.NetIncomeCommonStockholders.Scaled) / multiplier
		fmt.Printf("  Net Income: %.0f\n", actualValue)
	}
	if dto.Current.BasicEPS != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.BasicEPS.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.BasicEPS.Scaled) / multiplier
		fmt.Printf("  Basic EPS: %.2f %s\n", actualValue, dto.Currency)
	}
	if dto.Current.DilutedEPS != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.DilutedEPS.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.DilutedEPS.Scaled) / multiplier
		fmt.Printf("  Diluted EPS: %.2f %s\n", actualValue, dto.Currency)
	}
	if dto.Current.BasicAverageShares != nil {
		fmt.Printf("  Basic Average Shares: %d\n", *dto.Current.BasicAverageShares)
	}
	if dto.Current.DilutedAverageShares != nil {
		fmt.Printf("  Diluted Average Shares: %d\n", *dto.Current.DilutedAverageShares)
	}
	if dto.Current.TotalExpenses != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalExpenses.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalExpenses.Scaled) / multiplier
		fmt.Printf("  Total Expenses: %.0f\n", actualValue)
	}
	if dto.Current.EBIT != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EBIT.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EBIT.Scaled) / multiplier
		fmt.Printf("  EBIT: %.0f\n", actualValue)
	}
	if dto.Current.EBITDA != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EBITDA.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EBITDA.Scaled) / multiplier
		fmt.Printf("  EBITDA: %.0f\n", actualValue)
	}
	if dto.Current.NormalizedEBITDA != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.NormalizedEBITDA.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.NormalizedEBITDA.Scaled) / multiplier
		fmt.Printf("  Normalized EBITDA: %.0f\n", actualValue)
	}

	// Balance Sheet values
	fmt.Printf("\nBALANCE SHEET:\n")
	if dto.Current.TotalAssets != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalAssets.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalAssets.Scaled) / multiplier
		fmt.Printf("  Total Assets: %.0f\n", actualValue)
	}
	if dto.Current.TotalCapitalization != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalCapitalization.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalCapitalization.Scaled) / multiplier
		fmt.Printf("  Total Capitalization: %.0f\n", actualValue)
	}
	if dto.Current.CommonStockEquity != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.CommonStockEquity.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.CommonStockEquity.Scaled) / multiplier
		fmt.Printf("  Common Stock Equity: %.0f\n", actualValue)
	}
	if dto.Current.CapitalLeaseObligations != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.CapitalLeaseObligations.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.CapitalLeaseObligations.Scaled) / multiplier
		fmt.Printf("  Capital Lease Obligations: %.0f\n", actualValue)
	}
	if dto.Current.NetTangibleAssets != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.NetTangibleAssets.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.NetTangibleAssets.Scaled) / multiplier
		fmt.Printf("  Net Tangible Assets: %.0f\n", actualValue)
	}
	if dto.Current.WorkingCapital != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.WorkingCapital.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.WorkingCapital.Scaled) / multiplier
		fmt.Printf("  Working Capital: %.0f\n", actualValue)
	}
	if dto.Current.InvestedCapital != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.InvestedCapital.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.InvestedCapital.Scaled) / multiplier
		fmt.Printf("  Invested Capital: %.0f\n", actualValue)
	}
	if dto.Current.TangibleBookValue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TangibleBookValue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TangibleBookValue.Scaled) / multiplier
		fmt.Printf("  Tangible Book Value: %.0f\n", actualValue)
	}
	if dto.Current.TotalDebt != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.TotalDebt.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.TotalDebt.Scaled) / multiplier
		fmt.Printf("  Total Debt: %.0f\n", actualValue)
	}
	if dto.Current.ShareIssued != nil {
		fmt.Printf("  Share Issued: %d\n", *dto.Current.ShareIssued)
	}

	// Cash Flow values
	fmt.Printf("\nCASH FLOW:\n")
	if dto.Current.OperatingCashFlow != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.OperatingCashFlow.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.OperatingCashFlow.Scaled) / multiplier
		fmt.Printf("  Operating Cash Flow: %.0f\n", actualValue)
	}
	if dto.Current.InvestingCashFlow != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.InvestingCashFlow.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.InvestingCashFlow.Scaled) / multiplier
		fmt.Printf("  Investing Cash Flow: %.0f\n", actualValue)
	}
	if dto.Current.FinancingCashFlow != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.FinancingCashFlow.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.FinancingCashFlow.Scaled) / multiplier
		fmt.Printf("  Financing Cash Flow: %.0f\n", actualValue)
	}
	if dto.Current.EndCashPosition != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.EndCashPosition.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.EndCashPosition.Scaled) / multiplier
		fmt.Printf("  End Cash Position: %.0f\n", actualValue)
	}
	if dto.Current.CapitalExpenditure != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.CapitalExpenditure.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.CapitalExpenditure.Scaled) / multiplier
		fmt.Printf("  Capital Expenditure: %.0f\n", actualValue)
	}
	if dto.Current.IssuanceOfDebt != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.IssuanceOfDebt.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.IssuanceOfDebt.Scaled) / multiplier
		fmt.Printf("  Issuance of Debt: %.0f\n", actualValue)
	}
	if dto.Current.RepaymentOfDebt != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.RepaymentOfDebt.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.RepaymentOfDebt.Scaled) / multiplier
		fmt.Printf("  Repayment of Debt: %.0f\n", actualValue)
	}
	if dto.Current.RepurchaseOfCapitalStock != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.RepurchaseOfCapitalStock.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.RepurchaseOfCapitalStock.Scaled) / multiplier
		fmt.Printf("  Repurchase of Capital Stock: %.0f\n", actualValue)
	}
	if dto.Current.FreeCashFlow != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Current.FreeCashFlow.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Current.FreeCashFlow.Scaled) / multiplier
		fmt.Printf("  Free Cash Flow: %.0f\n", actualValue)
	}

	// Historical values
	fmt.Printf("HISTORICAL VALUES:\n")
	if dto.Historical.Q2_2025.TotalRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Historical.Q2_2025.TotalRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Historical.Q2_2025.TotalRevenue.Scaled) / multiplier
		fmt.Printf("  Q2 2025 Revenue: %.0f\n", actualValue)
	}
	if dto.Historical.Q1_2025.TotalRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Historical.Q1_2025.TotalRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Historical.Q1_2025.TotalRevenue.Scaled) / multiplier
		fmt.Printf("  Q1 2025 Revenue: %.0f\n", actualValue)
	}
	if dto.Historical.Q4_2024.TotalRevenue != nil {
		multiplier := float64(1)
		for i := 0; i < dto.Historical.Q4_2024.TotalRevenue.Scale; i++ {
			multiplier *= 10
		}
		actualValue := float64(dto.Historical.Q4_2024.TotalRevenue.Scaled) / multiplier
		fmt.Printf("  Q4 2024 Revenue: %.0f\n", actualValue)
	}

	fmt.Printf("EXTRACTED: %d fields\n", countFinancialsFields(dto))
}

// countFinancialsFields counts the number of extracted fields in financials data
func countFinancialsFields(dto *scrape.ComprehensiveFinancialsDTO) int {
	count := 0

	// Count current fields
	if dto.Current.TotalRevenue != nil {
		count++
	}
	if dto.Current.CostOfRevenue != nil {
		count++
	}
	if dto.Current.GrossProfit != nil {
		count++
	}
	if dto.Current.OperatingIncome != nil {
		count++
	}
	if dto.Current.NetIncomeCommonStockholders != nil {
		count++
	}
	if dto.Current.BasicEPS != nil {
		count++
	}
	if dto.Current.DilutedEPS != nil {
		count++
	}
	if dto.Current.EBITDA != nil {
		count++
	}

	// Count historical fields
	if dto.Historical.Q2_2025.TotalRevenue != nil {
		count++
	}
	if dto.Historical.Q1_2025.TotalRevenue != nil {
		count++
	}
	if dto.Historical.Q4_2024.TotalRevenue != nil {
		count++
	}

	return count
}
