// statistics.go — ComprehensiveKeyStatisticsDTO → stdout summary. Shared by
// `yfin comprehensive-stats` (cmd/fundamentals) and `yfin scrape
// --preview-json --endpoint key-statistics` (cmd/scrape), which previously
// each carried a byte-identical private copy.
package format

import (
	"fmt"

	"github.com/bizshuk/yfin/model"
)

// ComprehensiveStatistics prints a summary of comprehensive statistics
func ComprehensiveStatistics(dto *model.ComprehensiveKeyStatisticsDTO) {
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

// ComprehensiveProfile prints a summary of comprehensive profile
