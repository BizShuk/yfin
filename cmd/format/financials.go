// financials.go — ComprehensiveFinancialsDTO → stdout summary, covering the
// financials / balance-sheet / cash-flow pages, plus the populated-field
// counter it reports. Used by `yfin scrape --preview-json`.
package format

import (
	"fmt"

	"github.com/bizshuk/yfin/model"
)

func ComprehensiveFinancials(dto *model.ComprehensiveFinancialsDTO) {
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
func countFinancialsFields(dto *model.ComprehensiveFinancialsDTO) int {
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
