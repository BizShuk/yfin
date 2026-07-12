// scrape_dtos.go — scrape DTOs from svc/scrape: PeriodLine, Recommendation,
// QuarterlyEPS, Officer, Executive, KeyStatisticsDTO, FinancialsDTO,
// ProfileDTO, AnalysisDTO, AnalystInsightsDTO, plus all
// Comprehensive* versions (Finance/KeyStatistics/Profile/Analysis) and
// the raw scrape transport shape (YahooFinanceData, FinancialDataPoint).
//
// Originally split across svc/scrape/types_json.go + svc/scrape/financials.go
// + svc/scrape/profile.go + svc/scrape/statistics.go + svc/scrape/analysis.go
// + svc/scrape/analyst_insights.go. Consolidated here so any layer can
// reference these DTOs without pulling in svc/scrape.

package model

import "time"

// PeriodLine represents a financial statement line item for a specific period
type PeriodLine struct {
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Key         string    `json:"key"`
	Value       Scaled    `json:"value"`
	Currency    Currency  `json:"currency"`
}

// FinancialsDTO represents extracted financial statements data
type FinancialsDTO struct {
	Symbol string       `json:"symbol"`
	Market string       `json:"market"`
	Lines  []PeriodLine `json:"lines"`
	AsOf   time.Time    `json:"as_of"`
}

// Recommendation represents analyst recommendation data for a period
type Recommendation struct {
	Period     string `json:"period"`
	StrongBuy  int    `json:"strong_buy"`
	Buy        int    `json:"buy"`
	Hold       int    `json:"hold"`
	Sell       int    `json:"sell"`
	StrongSell int    `json:"strong_sell"`
}

// QuarterlyEPS represents quarterly EPS estimates and actuals
type QuarterlyEPS struct {
	Date     string  `json:"date"`
	Actual   *Scaled `json:"actual,omitempty"`
	Estimate *Scaled `json:"estimate,omitempty"`
}

// AnalysisDTO represents extracted analysis data
type AnalysisDTO struct {
	Symbol       string           `json:"symbol"`
	Market       string           `json:"market"`
	Currency     Currency         `json:"currency"`
	RecTrends    []Recommendation `json:"rec_trends"`
	EPSQuarterly []QuarterlyEPS   `json:"eps_quarterly"`
	AsOf         time.Time        `json:"as_of"`
}

// Officer represents a company officer/executive
type Officer struct {
	Name  string  `json:"name"`
	Title string  `json:"title"`
	Age   *int    `json:"age,omitempty"`
	Pay   *Scaled `json:"pay,omitempty"`
}

// ProfileDTO represents extracted company profile data
type ProfileDTO struct {
	Symbol    string    `json:"symbol"`
	Market    string    `json:"market"`
	Company   string    `json:"company"`
	Address1  string    `json:"address1"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	Country   string    `json:"country"`
	Phone     string    `json:"phone"`
	Website   string    `json:"website"`
	Industry  string    `json:"industry"`
	Sector    string    `json:"sector"`
	Employees *int      `json:"employees,omitempty"`
	Officers  []Officer `json:"officers"`
	AsOf      time.Time `json:"as_of"`
}

// KeyStatisticsDTO represents extracted key statistics data
type KeyStatisticsDTO struct {
	Symbol   string   `json:"symbol"`
	Market   string   `json:"market"`
	Currency Currency `json:"currency"`

	// Market metrics
	MarketCap    *Scaled `json:"market_cap,omitempty"`
	ForwardPE    *Scaled `json:"forward_pe,omitempty"`
	TrailingPE   *Scaled `json:"trailing_pe,omitempty"`
	Beta         *Scaled `json:"beta,omitempty"`
	PriceToSales *Scaled `json:"price_to_sales,omitempty"`

	// Share data
	SharesOutstanding *int64 `json:"shares_outstanding,omitempty"`
	FloatShares       *int64 `json:"float_shares,omitempty"`
	ShortInterest     *int64 `json:"short_interest,omitempty"`

	// Financial metrics
	EnterpriseValue  *Scaled `json:"enterprise_value,omitempty"`
	TotalCash        *Scaled `json:"total_cash,omitempty"`
	TotalDebt        *Scaled `json:"total_debt,omitempty"`
	QuickRatio       *Scaled `json:"quick_ratio,omitempty"`
	CurrentRatio     *Scaled `json:"current_ratio,omitempty"`
	DebtToEquity     *Scaled `json:"debt_to_equity,omitempty"`
	ReturnOnAssets   *Scaled `json:"return_on_assets,omitempty"`
	ReturnOnEquity   *Scaled `json:"return_on_equity,omitempty"`
	GrossMargins     *Scaled `json:"gross_margins,omitempty"`
	OperatingMargins *Scaled `json:"operating_margins,omitempty"`
	ProfitMargins    *Scaled `json:"profit_margins,omitempty"`
	RevenueGrowth    *Scaled `json:"revenue_growth,omitempty"`
	EarningsGrowth   *Scaled `json:"earnings_growth,omitempty"`

	// Price data
	FiftyTwoWeekHigh   *Scaled `json:"fifty_two_week_high,omitempty"`
	FiftyTwoWeekLow    *Scaled `json:"fifty_two_week_low,omitempty"`
	AverageVolume      *int64  `json:"average_volume,omitempty"`
	AverageVolume10Day *int64  `json:"average_volume_10_day,omitempty"`

	AsOf time.Time `json:"as_of"`
}

// Executive represents a company executive
type Executive struct {
	Name             string `json:"name,omitempty"`
	Title            string `json:"title,omitempty"`
	YearBorn         *int   `json:"year_born,omitempty"`
	TotalPay         *int64 `json:"total_pay,omitempty"`
	ExercisedValue   *int64 `json:"exercised_value,omitempty"`
	UnexercisedValue *int64 `json:"unexercised_value,omitempty"`
}

// ComprehensiveProfileDTO holds comprehensive profile data
type ComprehensiveProfileDTO struct {
	Symbol string    `json:"symbol"`
	Market string    `json:"market"`
	AsOf   time.Time `json:"as_of"`

	// Company Information
	CompanyName       string `json:"company_name,omitempty"`
	ShortName         string `json:"short_name,omitempty"`
	Address1          string `json:"address1,omitempty"`
	City              string `json:"city,omitempty"`
	State             string `json:"state,omitempty"`
	Zip               string `json:"zip,omitempty"`
	Country           string `json:"country,omitempty"`
	Phone             string `json:"phone,omitempty"`
	Website           string `json:"website,omitempty"`
	Industry          string `json:"industry,omitempty"`
	Sector            string `json:"sector,omitempty"`
	FullTimeEmployees *int64 `json:"full_time_employees,omitempty"`
	BusinessSummary   string `json:"business_summary,omitempty"`

	// Key Executives
	Executives []Executive `json:"executives,omitempty"`

	// Additional Information
	MaxAge                    *int64 `json:"max_age,omitempty"`
	AuditRisk                 *int64 `json:"audit_risk,omitempty"`
	BoardRisk                 *int64 `json:"board_risk,omitempty"`
	CompensationRisk          *int64 `json:"compensation_risk,omitempty"`
	ShareHolderRightsRisk     *int64 `json:"share_holder_rights_risk,omitempty"`
	OverallRisk               *int64 `json:"overall_risk,omitempty"`
	GovernanceEpochDate       *int64 `json:"governance_epoch_date,omitempty"`
	CompensationAsOfEpochDate *int64 `json:"compensation_as_of_epoch_date,omitempty"`
}

// AnalystInsightsDTO represents analyst insights data from Yahoo Finance
type AnalystInsightsDTO struct {
	Symbol string    `json:"symbol"`
	Market string    `json:"market"`
	AsOf   time.Time `json:"as_of"`

	// Price Targets
	CurrentPrice      *float64 `json:"current_price,omitempty"`
	TargetMeanPrice   *float64 `json:"target_mean_price,omitempty"`
	TargetMedianPrice *float64 `json:"target_median_price,omitempty"`
	TargetHighPrice   *float64 `json:"target_high_price,omitempty"`
	TargetLowPrice    *float64 `json:"target_low_price,omitempty"`

	// Analyst Opinions
	NumberOfAnalysts   *int     `json:"number_of_analysts,omitempty"`
	RecommendationMean *float64 `json:"recommendation_mean,omitempty"`
	RecommendationKey  *string  `json:"recommendation_key,omitempty"`
}

// ComprehensiveKeyStatisticsDTO holds all key statistics data
type ComprehensiveKeyStatisticsDTO struct {
	Symbol   string    `json:"symbol"`
	Market   string    `json:"market"`
	Currency string    `json:"currency"`
	AsOf     time.Time `json:"as_of"`

	// Current values (most recent data)
	Current struct {
		MarketCap              *Scaled `json:"market_cap,omitempty"`
		EnterpriseValue        *Scaled `json:"enterprise_value,omitempty"`
		TrailingPE             *Scaled `json:"trailing_pe,omitempty"`
		ForwardPE              *Scaled `json:"forward_pe,omitempty"`
		PEGRatio               *Scaled `json:"peg_ratio,omitempty"`
		PriceSales             *Scaled `json:"price_sales,omitempty"`
		PriceBook              *Scaled `json:"price_book,omitempty"`
		EnterpriseValueRevenue *Scaled `json:"enterprise_value_revenue,omitempty"`
		EnterpriseValueEBITDA  *Scaled `json:"enterprise_value_ebitda,omitempty"`
	} `json:"current"`

	// Additional statistics
	Additional struct {
		Beta              *Scaled `json:"beta,omitempty"`
		SharesOutstanding *int64  `json:"shares_outstanding,omitempty"`
		ProfitMargin      *Scaled `json:"profit_margin,omitempty"`
		OperatingMargin   *Scaled `json:"operating_margin,omitempty"`
		ReturnOnAssets    *Scaled `json:"return_on_assets,omitempty"`
		ReturnOnEquity    *Scaled `json:"return_on_equity,omitempty"`
	} `json:"additional"`

	// Historical values - dynamic quarters
	Historical []HistoricalQuarter `json:"historical,omitempty"`
}

// HistoricalQuarter represents one historical period's key statistics row
type HistoricalQuarter struct {
	Date                   string  `json:"date"`
	MarketCap              *Scaled `json:"market_cap,omitempty"`
	EnterpriseValue        *Scaled `json:"enterprise_value,omitempty"`
	TrailingPE             *Scaled `json:"trailing_pe,omitempty"`
	ForwardPE              *Scaled `json:"forward_pe,omitempty"`
	PEGRatio               *Scaled `json:"peg_ratio,omitempty"`
	PriceSales             *Scaled `json:"price_sales,omitempty"`
	PriceBook              *Scaled `json:"price_book,omitempty"`
	EnterpriseValueRevenue *Scaled `json:"enterprise_value_revenue,omitempty"`
	EnterpriseValueEBITDA  *Scaled `json:"enterprise_value_ebitda,omitempty"`
}

// FinancialDataPoint represents a single financial data point from Yahoo Finance
type FinancialDataPoint struct {
	DataID       int64  `json:"dataId"`
	AsOfDate     string `json:"asOfDate"`
	PeriodType   string `json:"periodType"`
	CurrencyCode string `json:"currencyCode"`
	ReportedValue struct {
		Raw float64 `json:"raw"`
		Fmt string  `json:"fmt"`
	} `json:"reportedValue"`
}

// YahooFinanceData represents the JSON structure from Yahoo Finance
type YahooFinanceData struct {
	QuoteSummary struct {
		Result []struct {
			FinancialData struct {
				TrailingTotalRevenue                         []FinancialDataPoint `json:"trailingTotalRevenue"`
				AnnualTotalRevenue                           []FinancialDataPoint `json:"annualTotalRevenue"`
				TrailingOperatingIncome                      []FinancialDataPoint `json:"trailingOperatingIncome"`
				AnnualOperatingIncome                        []FinancialDataPoint `json:"annualOperatingIncome"`
				TrailingNetIncome                            []FinancialDataPoint `json:"trailingNetIncome"`
				AnnualNetIncome                              []FinancialDataPoint `json:"annualNetIncome"`
				TrailingBasicEPS                             []FinancialDataPoint `json:"trailingBasicEPS"`
				AnnualBasicEPS                               []FinancialDataPoint `json:"annualBasicEPS"`
				TrailingDilutedEPS                           []FinancialDataPoint `json:"trailingDilutedEPS"`
				AnnualDilutedEPS                             []FinancialDataPoint `json:"annualDilutedEPS"`
				TrailingEBITDA                               []FinancialDataPoint `json:"trailingEBITDA"`
				AnnualEBITDA                                 []FinancialDataPoint `json:"annualEBITDA"`
				TrailingGrossProfit                          []FinancialDataPoint `json:"trailingGrossProfit"`
				AnnualGrossProfit                            []FinancialDataPoint `json:"annualGrossProfit"`
				TrailingCostOfRevenue                        []FinancialDataPoint `json:"trailingCostOfRevenue"`
				AnnualCostOfRevenue                          []FinancialDataPoint `json:"annualCostOfRevenue"`
				TrailingOperatingExpense                     []FinancialDataPoint `json:"trailingOperatingExpense"`
				AnnualOperatingExpense                       []FinancialDataPoint `json:"annualOperatingExpense"`
				TrailingTotalExpenses                        []FinancialDataPoint `json:"trailingTotalExpenses"`
				AnnualTotalExpenses                          []FinancialDataPoint `json:"annualTotalExpenses"`
				TrailingTaxProvision                         []FinancialDataPoint `json:"trailingTaxProvision"`
				AnnualTaxProvision                           []FinancialDataPoint `json:"annualTaxProvision"`
				TrailingPretaxIncome                         []FinancialDataPoint `json:"trailingPretaxIncome"`
				AnnualPretaxIncome                           []FinancialDataPoint `json:"annualPretaxIncome"`
				TrailingOtherIncomeExpense                   []FinancialDataPoint `json:"trailingOtherIncomeExpense"`
				AnnualOtherIncomeExpense                     []FinancialDataPoint `json:"annualOtherIncomeExpense"`
				TrailingNetNonOperatingInterestIncomeExpense []FinancialDataPoint `json:"trailingNetNonOperatingInterestIncomeExpense"`
				AnnualNetNonOperatingInterestIncomeExpense   []FinancialDataPoint `json:"annualNetNonOperatingInterestIncomeExpense"`
				TrailingBasicAverageShares                   []FinancialDataPoint `json:"trailingBasicAverageShares"`
				AnnualBasicAverageShares                     []FinancialDataPoint `json:"annualBasicAverageShares"`
				TrailingDilutedAverageShares                 []FinancialDataPoint `json:"trailingDilutedAverageShares"`
				AnnualDilutedAverageShares                   []FinancialDataPoint `json:"annualDilutedAverageShares"`
				TrailingEBIT                                 []FinancialDataPoint `json:"trailingEBIT"`
				AnnualEBIT                                   []FinancialDataPoint `json:"annualEBIT"`
				TrailingNormalizedIncome                     []FinancialDataPoint `json:"trailingNormalizedIncome"`
				AnnualNormalizedIncome                       []FinancialDataPoint `json:"annualNormalizedIncome"`
				TrailingNormalizedEBITDA                     []FinancialDataPoint `json:"trailingNormalizedEBITDA"`
				AnnualNormalizedEBITDA                       []FinancialDataPoint `json:"annualNormalizedEBITDA"`
				TrailingReconciledCostOfRevenue              []FinancialDataPoint `json:"trailingReconciledCostOfRevenue"`
				AnnualReconciledCostOfRevenue                []FinancialDataPoint `json:"annualReconciledCostOfRevenue"`
				TrailingReconciledDepreciation               []FinancialDataPoint `json:"trailingReconciledDepreciation"`
				AnnualReconciledDepreciation                 []FinancialDataPoint `json:"annualReconciledDepreciation"`
			} `json:"financialData"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

// ComprehensiveFinancialsDTO holds all financials data including historical
type ComprehensiveFinancialsDTO struct {
	Symbol   string    `json:"symbol"`
	Market   string    `json:"market"`
	Currency string    `json:"currency"`
	AsOf     time.Time `json:"as_of"`

	// Current values (most recent quarter)
	Current struct {
		TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
		CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
		GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
		OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
		OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
		NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
		OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
		PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
		TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
		NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
		BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
		DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
		BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
		DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
		TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
		NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
		EBIT                                 *Scaled `json:"ebit,omitempty"`
		EBITDA                               *Scaled `json:"ebitda,omitempty"`
		ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
		ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
		NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`

		// Balance Sheet fields
		TotalAssets             *Scaled `json:"total_assets,omitempty"`
		TotalCapitalization     *Scaled `json:"total_capitalization,omitempty"`
		CommonStockEquity       *Scaled `json:"common_stock_equity,omitempty"`
		CapitalLeaseObligations *Scaled `json:"capital_lease_obligations,omitempty"`
		NetTangibleAssets       *Scaled `json:"net_tangible_assets,omitempty"`
		WorkingCapital          *Scaled `json:"working_capital,omitempty"`
		InvestedCapital         *Scaled `json:"invested_capital,omitempty"`
		TangibleBookValue       *Scaled `json:"tangible_book_value,omitempty"`
		TotalDebt               *Scaled `json:"total_debt,omitempty"`
		ShareIssued             *int64  `json:"share_issued,omitempty"`

		// Cash Flow fields
		OperatingCashFlow        *Scaled `json:"operating_cash_flow,omitempty"`
		InvestingCashFlow        *Scaled `json:"investing_cash_flow,omitempty"`
		FinancingCashFlow        *Scaled `json:"financing_cash_flow,omitempty"`
		EndCashPosition          *Scaled `json:"end_cash_position,omitempty"`
		CapitalExpenditure       *Scaled `json:"capital_expenditure,omitempty"`
		IssuanceOfDebt           *Scaled `json:"issuance_of_debt,omitempty"`
		RepaymentOfDebt          *Scaled `json:"repayment_of_debt,omitempty"`
		RepurchaseOfCapitalStock *Scaled `json:"repurchase_of_capital_stock,omitempty"`
		FreeCashFlow             *Scaled `json:"free_cash_flow,omitempty"`
	} `json:"current"`

	// Historical values
	Historical struct {
		Q2_2025 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q2_2025"`

		Q1_2025 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q1_2025"`

		Q4_2024 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q4_2024"`

		Q3_2024 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q3_2024"`

		Q2_2024 struct {
			Date                                 string  `json:"date"`
			TotalRevenue                         *Scaled `json:"total_revenue,omitempty"`
			CostOfRevenue                        *Scaled `json:"cost_of_revenue,omitempty"`
			GrossProfit                          *Scaled `json:"gross_profit,omitempty"`
			OperatingExpense                     *Scaled `json:"operating_expense,omitempty"`
			OperatingIncome                      *Scaled `json:"operating_income,omitempty"`
			NetNonOperatingInterestIncomeExpense *Scaled `json:"net_non_operating_interest_income_expense,omitempty"`
			OtherIncomeExpense                   *Scaled `json:"other_income_expense,omitempty"`
			PretaxIncome                         *Scaled `json:"pretax_income,omitempty"`
			TaxProvision                         *Scaled `json:"tax_provision,omitempty"`
			NetIncomeCommonStockholders          *Scaled `json:"net_income_common_stockholders,omitempty"`
			BasicEPS                             *Scaled `json:"basic_eps,omitempty"`
			DilutedEPS                           *Scaled `json:"diluted_eps,omitempty"`
			BasicAverageShares                   *int64  `json:"basic_average_shares,omitempty"`
			DilutedAverageShares                 *int64  `json:"diluted_average_shares,omitempty"`
			TotalExpenses                        *Scaled `json:"total_expenses,omitempty"`
			NormalizedIncome                     *Scaled `json:"normalized_income,omitempty"`
			EBIT                                 *Scaled `json:"ebit,omitempty"`
			EBITDA                               *Scaled `json:"ebitda,omitempty"`
			ReconciledCostOfRevenue              *Scaled `json:"reconciled_cost_of_revenue,omitempty"`
			ReconciledDepreciation               *Scaled `json:"reconciled_depreciation,omitempty"`
			NormalizedEBITDA                     *Scaled `json:"normalized_ebitda,omitempty"`
		} `json:"q2_2024"`
	} `json:"historical"`
}

// ComprehensiveAnalysisDTO represents comprehensive analysis data from Yahoo Finance
type ComprehensiveAnalysisDTO struct {
	Symbol string    `json:"symbol"`
	Market string    `json:"market"`
	AsOf   time.Time `json:"as_of"`

	// Earnings Estimate
	EarningsEstimate struct {
		Currency   string `json:"currency"`
		CurrentQtr struct {
			NoOfAnalysts *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate  *float64 `json:"avg_estimate,omitempty"`
			LowEstimate  *float64 `json:"low_estimate,omitempty"`
			HighEstimate *float64 `json:"high_estimate,omitempty"`
			YearAgoEPS   *float64 `json:"year_ago_eps,omitempty"`
		} `json:"current_qtr"`
		NextQtr struct {
			NoOfAnalysts *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate  *float64 `json:"avg_estimate,omitempty"`
			LowEstimate  *float64 `json:"low_estimate,omitempty"`
			HighEstimate *float64 `json:"high_estimate,omitempty"`
			YearAgoEPS   *float64 `json:"year_ago_eps,omitempty"`
		} `json:"next_qtr"`
		CurrentYear struct {
			NoOfAnalysts *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate  *float64 `json:"avg_estimate,omitempty"`
			LowEstimate  *float64 `json:"low_estimate,omitempty"`
			HighEstimate *float64 `json:"high_estimate,omitempty"`
			YearAgoEPS   *float64 `json:"year_ago_eps,omitempty"`
		} `json:"current_year"`
		NextYear struct {
			NoOfAnalysts *int     `json:"no_of_analysts,omitempty"`
			AvgEstimate  *float64 `json:"avg_estimate,omitempty"`
			LowEstimate  *float64 `json:"low_estimate,omitempty"`
			HighEstimate *float64 `json:"high_estimate,omitempty"`
			YearAgoEPS   *float64 `json:"year_ago_eps,omitempty"`
		} `json:"next_year"`
	} `json:"earnings_estimate"`

	// Revenue Estimate
	RevenueEstimate struct {
		Currency   string `json:"currency"`
		CurrentQtr struct {
			NoOfAnalysts       *int    `json:"no_of_analysts,omitempty"`
			AvgEstimate        *string `json:"avg_estimate,omitempty"`
			LowEstimate        *string `json:"low_estimate,omitempty"`
			HighEstimate       *string `json:"high_estimate,omitempty"`
			YearAgoSales       *string `json:"year_ago_sales,omitempty"`
			SalesGrowthYearEst *string `json:"sales_growth_year_est,omitempty"`
		} `json:"current_qtr"`
		NextQtr struct {
			NoOfAnalysts       *int    `json:"no_of_analysts,omitempty"`
			AvgEstimate        *string `json:"avg_estimate,omitempty"`
			LowEstimate        *string `json:"low_estimate,omitempty"`
			HighEstimate       *string `json:"high_estimate,omitempty"`
			YearAgoSales       *string `json:"year_ago_sales,omitempty"`
			SalesGrowthYearEst *string `json:"sales_growth_year_est,omitempty"`
		} `json:"next_qtr"`
		CurrentYear struct {
			NoOfAnalysts       *int    `json:"no_of_analysts,omitempty"`
			AvgEstimate        *string `json:"avg_estimate,omitempty"`
			LowEstimate        *string `json:"low_estimate,omitempty"`
			HighEstimate       *string `json:"high_estimate,omitempty"`
			YearAgoSales       *string `json:"year_ago_sales,omitempty"`
			SalesGrowthYearEst *string `json:"sales_growth_year_est,omitempty"`
		} `json:"current_year"`
		NextYear struct {
			NoOfAnalysts       *int    `json:"no_of_analysts,omitempty"`
			AvgEstimate        *string `json:"avg_estimate,omitempty"`
			LowEstimate        *string `json:"low_estimate,omitempty"`
			HighEstimate       *string `json:"high_estimate,omitempty"`
			YearAgoSales       *string `json:"year_ago_sales,omitempty"`
			SalesGrowthYearEst *string `json:"sales_growth_year_est,omitempty"`
		} `json:"next_year"`
	} `json:"revenue_estimate"`

	// Earnings History (dynamic dates)
	EarningsHistory struct {
		Currency string `json:"currency"`
		Data     []struct {
			Date            string   `json:"date"`
			EPSEst          *float64 `json:"eps_est,omitempty"`
			EPSActual       *float64 `json:"eps_actual,omitempty"`
			Difference      *float64 `json:"difference,omitempty"`
			SurprisePercent *string  `json:"surprise_percent,omitempty"`
		} `json:"data"`
	} `json:"earnings_history"`

	// EPS Trend
	EPSTrend struct {
		Currency   string `json:"currency"`
		CurrentQtr struct {
			CurrentEstimate *float64 `json:"current_estimate,omitempty"`
			Days7Ago        *float64 `json:"days_7_ago,omitempty"`
			Days30Ago       *float64 `json:"days_30_ago,omitempty"`
			Days60Ago       *float64 `json:"days_60_ago,omitempty"`
			Days90Ago       *float64 `json:"days_90_ago,omitempty"`
		} `json:"current_qtr"`
		NextQtr struct {
			CurrentEstimate *float64 `json:"current_estimate,omitempty"`
			Days7Ago        *float64 `json:"days_7_ago,omitempty"`
			Days30Ago       *float64 `json:"days_30_ago,omitempty"`
			Days60Ago       *float64 `json:"days_60_ago,omitempty"`
			Days90Ago       *float64 `json:"days_90_ago,omitempty"`
		} `json:"next_qtr"`
		CurrentYear struct {
			CurrentEstimate *float64 `json:"current_estimate,omitempty"`
			Days7Ago        *float64 `json:"days_7_ago,omitempty"`
			Days30Ago       *float64 `json:"days_30_ago,omitempty"`
			Days60Ago       *float64 `json:"days_60_ago,omitempty"`
			Days90Ago       *float64 `json:"days_90_ago,omitempty"`
		} `json:"current_year"`
		NextYear struct {
			CurrentEstimate *float64 `json:"current_estimate,omitempty"`
			Days7Ago        *float64 `json:"days_7_ago,omitempty"`
			Days30Ago       *float64 `json:"days_30_ago,omitempty"`
			Days60Ago       *float64 `json:"days_60_ago,omitempty"`
			Days90Ago       *float64 `json:"days_90_ago,omitempty"`
		} `json:"next_year"`
	} `json:"eps_trend"`

	// EPS Revisions
	EPSRevisions struct {
		Currency   string `json:"currency"`
		CurrentQtr struct {
			UpLast7Days    *int `json:"up_last_7_days,omitempty"`
			UpLast30Days   *int `json:"up_last_30_days,omitempty"`
			DownLast7Days  *int `json:"down_last_7_days,omitempty"`
			DownLast30Days *int `json:"down_last_30_days,omitempty"`
		} `json:"current_qtr"`
		NextQtr struct {
			UpLast7Days    *int `json:"up_last_7_days,omitempty"`
			UpLast30Days   *int `json:"up_last_30_days,omitempty"`
			DownLast7Days  *int `json:"down_last_7_days,omitempty"`
			DownLast30Days *int `json:"down_last_30_days,omitempty"`
		} `json:"next_qtr"`
		CurrentYear struct {
			UpLast7Days    *int `json:"up_last_7_days,omitempty"`
			UpLast30Days   *int `json:"up_last_30_days,omitempty"`
			DownLast7Days  *int `json:"down_last_7_days,omitempty"`
			DownLast30Days *int `json:"down_last_30_days,omitempty"`
		} `json:"current_year"`
		NextYear struct {
			UpLast7Days    *int `json:"up_last_7_days,omitempty"`
			UpLast30Days   *int `json:"up_last_30_days,omitempty"`
			DownLast7Days  *int `json:"down_last_7_days,omitempty"`
			DownLast30Days *int `json:"down_last_30_days,omitempty"`
		} `json:"next_year"`
	} `json:"eps_revisions"`

	// Growth Estimates (only ticker data, not S&P 500)
	GrowthEstimate struct {
		CurrentQtr  *string `json:"current_qtr,omitempty"`
		NextQtr     *string `json:"next_qtr,omitempty"`
		CurrentYear *string `json:"current_year,omitempty"`
		NextYear    *string `json:"next_year,omitempty"`
	} `json:"growth_estimate"`
}