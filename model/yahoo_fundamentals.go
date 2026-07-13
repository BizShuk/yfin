// yahoo_fundamentals.go — raw Yahoo Finance quarterly fundamentals structs.
// Originally lived in svc/yahoo/fundamentals.go; promoted to model/ so external
// consumers (cmd, facade) can depend on the raw API shape without importing
// the Decode/Validate/Fetch behavior of svc/yahoo.

package model

import "fmt"

// FundamentalsResponse represents the Yahoo Finance fundamentals API response
type FundamentalsResponse struct {
	QuoteSummary QuoteSummary `json:"quoteSummary"`
}

// QuoteSummary contains the fundamentals data
type QuoteSummary struct {
	Result []FundamentalsResult `json:"result"`
	Error  *string              `json:"error"`
}

// FundamentalsResult contains fundamentals data for a single symbol
type FundamentalsResult struct {
	IncomeStatementHistoryQuarterly   *IncomeStatementHistory   `json:"incomeStatementHistoryQuarterly"`
	BalanceSheetHistoryQuarterly      *BalanceSheetHistory      `json:"balanceSheetHistoryQuarterly"`
	CashflowStatementHistoryQuarterly *CashflowStatementHistory `json:"cashflowStatementHistoryQuarterly"`
}

// IncomeStatementHistory contains quarterly income statement data
type IncomeStatementHistory struct {
	IncomeStatementHistory []IncomeStatement `json:"incomeStatementHistory"`
}

// BalanceSheetHistory contains quarterly balance sheet data
type BalanceSheetHistory struct {
	BalanceSheetHistory []BalanceSheet `json:"balanceSheetHistory"`
}

// CashflowStatementHistory contains quarterly cashflow statement data
type CashflowStatementHistory struct {
	CashflowStatementHistory []CashflowStatement `json:"cashflowStatementHistory"`
}

// IncomeStatement represents a single income statement
type IncomeStatement struct {
	MaxAge                       int64     `json:"maxAge"`
	EndDate                      DateValue `json:"endDate"`
	TotalRevenue                 *Value    `json:"totalRevenue"`
	CostOfRevenue                *Value    `json:"costOfRevenue"`
	GrossProfit                  *Value    `json:"grossProfit"`
	ResearchDevelopment          *Value    `json:"researchDevelopment"`
	SellingGeneralAdministrative *Value    `json:"sellingGeneralAdministrative"`
	TotalOperatingExpenses       *Value    `json:"totalOperatingExpenses"`
	OperatingIncome              *Value    `json:"operatingIncome"`
	TotalOtherIncomeExpenseNet   *Value    `json:"totalOtherIncomeExpenseNet"`
	EBIT                         *Value    `json:"ebit"`
	InterestExpense              *Value    `json:"interestExpense"`
	IncomeBeforeTax              *Value    `json:"incomeBeforeTax"`
	IncomeTaxExpense             *Value    `json:"incomeTaxExpense"`
	NetIncome                    *Value    `json:"netIncome"`
	NetIncomeCommonStockholders  *Value    `json:"netIncomeCommonStockholders"`
	EPS                          *Value    `json:"eps"`
	EPSDiluted                   *Value    `json:"epsDiluted"`
	WeightedAverageShares        *Value    `json:"weightedAverageShares"`
	WeightedAverageSharesDiluted *Value    `json:"weightedAverageSharesDiluted"`
}

// BalanceSheet represents a single balance sheet
type BalanceSheet struct {
	MaxAge                  int64     `json:"maxAge"`
	EndDate                 DateValue `json:"endDate"`
	Cash                    *Value    `json:"cash"`
	ShortTermInvestments    *Value    `json:"shortTermInvestments"`
	NetReceivables          *Value    `json:"netReceivables"`
	Inventory               *Value    `json:"inventory"`
	OtherCurrentAssets      *Value    `json:"otherCurrentAssets"`
	TotalCurrentAssets      *Value    `json:"totalCurrentAssets"`
	LongTermInvestments     *Value    `json:"longTermInvestments"`
	PropertyPlantEquipment  *Value    `json:"propertyPlantEquipment"`
	OtherAssets             *Value    `json:"otherAssets"`
	TotalAssets             *Value    `json:"totalAssets"`
	AccountsPayable         *Value    `json:"accountsPayable"`
	ShortLongTermDebt       *Value    `json:"shortLongTermDebt"`
	OtherCurrentLiab        *Value    `json:"otherCurrentLiab"`
	LongTermDebt            *Value    `json:"longTermDebt"`
	OtherLiab               *Value    `json:"otherLiab"`
	TotalCurrentLiabilities *Value    `json:"totalCurrentLiabilities"`
	TotalLiab               *Value    `json:"totalLiab"`
	CommonStock             *Value    `json:"commonStock"`
	RetainedEarnings        *Value    `json:"retainedEarnings"`
	TreasuryStock           *Value    `json:"treasuryStock"`
	OtherStockholderEquity  *Value    `json:"otherStockholderEquity"`
	TotalStockholderEquity  *Value    `json:"totalStockholderEquity"`
	NetTangibleAssets       *Value    `json:"netTangibleAssets"`
}

// CashflowStatement represents a single cashflow statement
type CashflowStatement struct {
	MaxAge                                int64     `json:"maxAge"`
	EndDate                               DateValue `json:"endDate"`
	Investments                           *Value    `json:"investments"`
	ChangeToLiabilities                   *Value    `json:"changeToLiabilities"`
	TotalCashflowsFromInvestingActivities *Value    `json:"totalCashflowsFromInvestingActivities"`
	NetBorrowings                         *Value    `json:"netBorrowings"`
	TotalCashFromFinancingActivities      *Value    `json:"totalCashFromFinancingActivities"`
	ChangeToOperatingActivities           *Value    `json:"changeToOperatingActivities"`
	NetIncome                             *Value    `json:"netIncome"`
	ChangeInCash                          *Value    `json:"changeInCash"`
	BeginPeriodCashFlow                   *Value    `json:"beginPeriodCashFlow"`
	EndPeriodCashFlow                     *Value    `json:"endPeriodCashFlow"`
	TotalCashFromOperatingActivities      *Value    `json:"totalCashFromOperatingActivities"`
	Depreciation                          *Value    `json:"depreciation"`
	OtherCashflowsFromInvestingActivities *Value    `json:"otherCashflowsFromInvestingActivities"`
	DividendsPaid                         *Value    `json:"dividendsPaid"`
	ChangeToInventory                     *Value    `json:"changeToInventory"`
	ChangeToAccountReceivables            *Value    `json:"changeToAccountReceivables"`
	SalePurchaseOfStock                   *Value    `json:"salePurchaseOfStock"`
	OtherCashflowsFromFinancingActivities *Value    `json:"otherCashflowsFromFinancingActivities"`
	ChangeToNetincome                     *Value    `json:"changeToNetincome"`
	CapitalExpenditures                   *Value    `json:"capitalExpenditures"`
	ChangeReceivables                     *Value    `json:"changeReceivables"`
	CashFlowsOtherOperating               *Value    `json:"cashFlowsOtherOperating"`
	ExchangeRateChanges                   *Value    `json:"exchangeRateChanges"`
	CashAndCashEquivalentsChanges         *Value    `json:"cashAndCashEquivalentsChanges"`
	ChangeInWorkingCapital                *Value    `json:"changeInWorkingCapital"`
}

// DateValue represents a date with raw timestamp and formatted string
type DateValue struct {
	Raw int64  `json:"raw"`
	Fmt string `json:"fmt"`
}

// Value represents a financial value with raw number and formatted string
type Value struct {
	Raw     *int64  `json:"raw"`
	Fmt     *string `json:"fmt"`
	LongFmt *string `json:"longFmt"`
}

// Fundamentals contains all financial statement data
type Fundamentals struct {
	IncomeStatements   []IncomeStatement   `json:"incomeStatements,omitempty"`
	BalanceSheets      []BalanceSheet      `json:"balanceSheets,omitempty"`
	CashflowStatements []CashflowStatement `json:"cashflowStatements,omitempty"`
}

// GetFundamentals extracts fundamentals data from the response
func (r *FundamentalsResponse) GetFundamentals() (*Fundamentals, error) {
	if len(r.QuoteSummary.Result) == 0 {
		return nil, fmt.Errorf("no fundamentals results")
	}

	result := r.QuoteSummary.Result[0]
	fundamentals := &Fundamentals{}

	// Extract income statement data
	if result.IncomeStatementHistoryQuarterly != nil {
		fundamentals.IncomeStatements = result.IncomeStatementHistoryQuarterly.IncomeStatementHistory
	}

	// Extract balance sheet data
	if result.BalanceSheetHistoryQuarterly != nil {
		fundamentals.BalanceSheets = result.BalanceSheetHistoryQuarterly.BalanceSheetHistory
	}

	// Extract cashflow statement data
	if result.CashflowStatementHistoryQuarterly != nil {
		fundamentals.CashflowStatements = result.CashflowStatementHistoryQuarterly.CashflowStatementHistory
	}

	return fundamentals, nil
}