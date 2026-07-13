// yahoo_raw.go — Raw Yahoo Finance HTTP response shapes that
// model/normalize.go consumes as inputs. Owned by `model/` (the lowest
// layer) rather than `svc/yahoo` because:
//
//  1. `svc/yahoo` does HTTP fetching and JSON decoding — its job is to
//     hand back bytes + envelope-level errors. The raw shape structs are
//     pure data, not behavior.
//  2. `model/normalize.go` consumes these shapes to produce
//     `model.Normalized*` types. If raw shapes lived in `svc/yahoo`,
//     `model/` would have to import `svc/yahoo` (correct direction).
//     But then any decoder that wants to return `model.ChartBar`
//     from `svc/yahoo` would force `svc/yahoo → model/`, creating a
//     cycle. By owning the raw shapes in `model/`, decoders in
//     `svc/yahoo` return `*model.BarsResponse` etc. directly, with no
//     reverse dependency.
//
// Per the architectural rule recorded in CLAUDE.md "Conventions 6":
// `model/` is the lowest layer; `svc/yahoo` may import `model/`.
// External SDK objects get decoded at the upper layer (facade / handler),
// not buried inside `model/`.
//
// Naming note: prefix `Chart` / `Quote` / `Fundamentals` avoids collision
// with the SDK-surface `model.Bar` / `model.Quote` / `model.CompanyInfo`
// which carry post-normalized (float64, currency-resolved) shapes.

package model

// ChartResponse is the Yahoo Finance `/v8/finance/chart` response envelope.
type ChartResponse struct {
	Chart Chart `json:"chart"`
}

// Chart is the inner envelope for chart data.
type Chart struct {
	Result []ChartResult `json:"result"`
	Error  *string       `json:"error"`
}

// ChartResult holds the per-symbol chart payload.
type ChartResult struct {
	Meta       ChartMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators ChartIndicators `json:"indicators"`
}

// ChartMeta is the chart metadata block.
type ChartMeta struct {
	Currency             string                `json:"currency"`
	Symbol               string                `json:"symbol"`
	ExchangeName         string                `json:"exchangeName"`
	FullExchangeName     string                `json:"fullExchangeName"`
	InstrumentType       string                `json:"instrumentType"`
	FirstTradeDate       int64                 `json:"firstTradeDate"`
	RegularMarketTime    int64                 `json:"regularMarketTime"`
	HasPrePostMarketData bool                  `json:"hasPrePostMarketData"`
	GmtOffset            int64                 `json:"gmtoffset"`
	Timezone             string                `json:"timezone"`
	ExchangeTimezoneName string                `json:"exchangeTimezoneName"`
	RegularMarketPrice   *float64              `json:"regularMarketPrice"`
	FiftyTwoWeekHigh     *float64              `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow      *float64              `json:"fiftyTwoWeekLow"`
	RegularMarketDayHigh *float64              `json:"regularMarketDayHigh"`
	RegularMarketDayLow  *float64              `json:"regularMarketDayLow"`
	RegularMarketVolume  *int64                `json:"regularMarketVolume"`
	LongName             string                `json:"longName"`
	ShortName            string                `json:"shortName"`
	ChartPreviousClose   *float64              `json:"chartPreviousClose"`
	PreviousClose        *float64              `json:"previousClose"`
	Scale                int                   `json:"scale"`
	PriceHint            int                   `json:"priceHint"`
	CurrentTradingPeriod *CurrentTradingPeriod `json:"currentTradingPeriod"`
	DataGranularity      string                `json:"dataGranularity"`
	Range                string                `json:"range"`
	ValidRanges          []string              `json:"validRanges"`
}

// CurrentTradingPeriod holds pre/regular/post trading windows.
type CurrentTradingPeriod struct {
	Pre     *TradingPeriod `json:"pre"`
	Regular *TradingPeriod `json:"regular"`
	Post    *TradingPeriod `json:"post"`
}

// TradingPeriod is one trading window.
type TradingPeriod struct {
	Timezone  string `json:"timezone"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	GmtOffset int64  `json:"gmtoffset"`
}

// ChartIndicators holds price + volume + adjusted-close indicator arrays.
type ChartIndicators struct {
	Quote    []QuoteIndicator    `json:"quote"`
	AdjClose []AdjCloseIndicator `json:"adjclose"`
}

// QuoteIndicator holds the per-timestamp OHLCV indicator arrays.
type QuoteIndicator struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

// AdjCloseIndicator holds adjusted-close prices.
type AdjCloseIndicator struct {
	AdjClose []*float64 `json:"adjclose"`
}

// ChartBar is one OHLCV bar from the chart response (raw shape).
// Renamed from `yahoo.Bar` to avoid colliding with the SDK-surface
// `model.Bar` (which is post-decode + currency-resolved).
type ChartBar struct {
	Timestamp int64    `json:"timestamp"`
	Open      float64  `json:"open"`
	High      float64  `json:"high"`
	Low       float64  `json:"low"`
	Close     float64  `json:"close"`
	Volume    int64    `json:"volume"`
	AdjClose  *float64 `json:"adjclose,omitempty"`
}

// QuoteResponse is the Yahoo Finance quote API response envelope.
type QuoteResponse struct {
	QuoteResponse QuoteResponseData `json:"quoteResponse"`
}

// QuoteResponseData wraps the inner result slice.
type QuoteResponseData struct {
	Result []QuoteResult `json:"result"`
	Error  *string       `json:"error"`
}

// QuoteResult holds one quote entry.
type QuoteResult struct {
	Language                   string   `json:"language"`
	Region                     string   `json:"region"`
	QuoteType                  string   `json:"quoteType"`
	TypeDisp                   string   `json:"typeDisp"`
	QuoteSourceName            string   `json:"quoteSourceName"`
	Triggerable                bool     `json:"triggerable"`
	CustomPriceAlertConfidence string   `json:"customPriceAlertConfidence"`
	Currency                   string   `json:"currency"`
	Exchange                   string   `json:"exchange"`
	ShortName                  string   `json:"shortName"`
	LongName                   string   `json:"longName"`
	MessageBoardId             string   `json:"messageBoardId"`
	ExchangeTimezoneName       string   `json:"exchangeTimezoneName"`
	ExchangeTimezoneShortName  string   `json:"exchangeTimezoneShortName"`
	GmtOffsetMilliseconds      int64    `json:"gmtOffSetMilliseconds"`
	Market                     string   `json:"market"`
	EsgPopulated               bool     `json:"esgPopulated"`
	RegularMarketPrice         *float64 `json:"regularMarketPrice"`
	RegularMarketTime          *int64   `json:"regularMarketTime"`
	RegularMarketChange        *float64 `json:"regularMarketChange"`
	RegularMarketOpen          *float64 `json:"regularMarketOpen"`
	RegularMarketDayHigh       *float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow        *float64 `json:"regularMarketDayLow"`
	RegularMarketVolume        *int64   `json:"regularMarketVolume"`
	Bid                        *float64 `json:"bid"`
	Ask                        *float64 `json:"ask"`
	BidSize                    *int64   `json:"bidSize"`
	AskSize                    *int64   `json:"askSize"`
	FullExchangeName           string   `json:"fullExchangeName"`
	FinancialCurrency          string   `json:"financialCurrency"`
	RegularMarketChangePercent *float64 `json:"regularMarketChangePercent"`
	MarketState                string   `json:"marketState"`
	Symbol                     string   `json:"symbol"`
}

// RawQuote is one quote (raw shape). Renamed from `yahoo.Quote` to
// avoid colliding with the SDK-surface `model.Quote`.
type RawQuote struct {
	Symbol                     string   `json:"symbol"`
	Currency                   string   `json:"currency"`
	Exchange                   string   `json:"exchange"`
	FullExchangeName           string   `json:"fullExchangeName"`
	ShortName                  string   `json:"shortName"`
	LongName                   string   `json:"longName"`
	QuoteType                  string   `json:"quoteType"`
	MarketState                string   `json:"marketState"`
	ExchangeTimezoneName       string   `json:"exchangeTimezoneName"`
	ExchangeTimezoneShortName  string   `json:"exchangeTimezoneShortName"`
	GmtOffsetMilliseconds      int64    `json:"gmtOffsetMilliseconds"`
	Bid                        *float64 `json:"bid,omitempty"`
	Ask                        *float64 `json:"ask,omitempty"`
	BidSize                    *int64   `json:"bidSize,omitempty"`
	AskSize                    *int64   `json:"askSize,omitempty"`
	RegularMarketPrice         *float64 `json:"regularMarketPrice,omitempty"`
	RegularMarketTime          *int64   `json:"regularMarketTime,omitempty"`
	RegularMarketChange        *float64 `json:"regularMarketChange,omitempty"`
	RegularMarketOpen          *float64 `json:"regularMarketOpen,omitempty"`
	RegularMarketDayHigh       *float64 `json:"regularMarketDayHigh,omitempty"`
	RegularMarketDayLow        *float64 `json:"regularMarketDayLow,omitempty"`
	RegularMarketVolume        *int64   `json:"regularMarketVolume,omitempty"`
	RegularMarketChangePercent *float64 `json:"regularMarketChangePercent,omitempty"`
}

// FundamentalsResponse is the Yahoo Finance fundamentals API response.
type FundamentalsResponse struct {
	QuoteSummary QuoteSummary `json:"quoteSummary"`
}

// QuoteSummary wraps the inner fundamentals results.
type QuoteSummary struct {
	Result []FundamentalsResult `json:"result"`
	Error  *string              `json:"error"`
}

// FundamentalsResult holds one fundamentals payload (per symbol).
type FundamentalsResult struct {
	IncomeStatementHistoryQuarterly   *IncomeStatementHistory   `json:"incomeStatementHistoryQuarterly"`
	BalanceSheetHistoryQuarterly      *BalanceSheetHistory      `json:"balanceSheetHistoryQuarterly"`
	CashflowStatementHistoryQuarterly *CashflowStatementHistory `json:"cashflowStatementHistoryQuarterly"`
}

// IncomeStatementHistory wraps a slice of income statements.
type IncomeStatementHistory struct {
	IncomeStatementHistory []IncomeStatement `json:"incomeStatementHistory"`
}

// BalanceSheetHistory wraps a slice of balance sheets.
type BalanceSheetHistory struct {
	BalanceSheetHistory []BalanceSheet `json:"balanceSheetHistory"`
}

// CashflowStatementHistory wraps a slice of cashflow statements.
type CashflowStatementHistory struct {
	CashflowStatementHistory []CashflowStatement `json:"cashflowStatementHistory"`
}

// IncomeStatement represents one quarter's income statement.
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

// BalanceSheet represents one quarter's balance sheet.
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
	LongTermDebt            *Value    `json:"longLongTermDebt"`
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

// CashflowStatement represents one quarter's cashflow statement.
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

// DateValue is a Yahoo date pair (raw epoch + formatted string).
type DateValue struct {
	Raw int64  `json:"raw"`
	Fmt string `json:"fmt"`
}

// Value is a Yahoo financial value triple (raw int64 + display fmt).
type Value struct {
	Raw     *int64  `json:"raw"`
	Fmt     *string `json:"fmt"`
	LongFmt *string `json:"longFmt"`
}

// Fundamentals is the aggregated shape after GetFundamentals() pulls
// the three statement histories into flat slices.
type Fundamentals struct {
	IncomeStatements   []IncomeStatement   `json:"incomeStatements,omitempty"`
	BalanceSheets      []BalanceSheet      `json:"balanceSheets,omitempty"`
	CashflowStatements []CashflowStatement `json:"cashflowStatements,omitempty"`
}

// GetBars extracts per-symbol bars from a chart response.
func (r *ChartResponse) GetBars() ([]ChartBar, error) {
	if len(r.Chart.Result) == 0 {
		return nil, ErrNoChartResults
	}
	result := r.Chart.Result[0]
	if len(result.Timestamp) == 0 || len(result.Indicators.Quote) == 0 {
		return []ChartBar{}, nil
	}
	quote := result.Indicators.Quote[0]
	bars := make([]ChartBar, 0, len(result.Timestamp))
	for i, ts := range result.Timestamp {
		if quote.Open[i] == nil || quote.High[i] == nil || quote.Low[i] == nil || quote.Close[i] == nil || quote.Volume[i] == nil {
			continue
		}
		bar := ChartBar{
			Timestamp: ts,
			Open:      *quote.Open[i],
			High:      *quote.High[i],
			Low:       *quote.Low[i],
			Close:     *quote.Close[i],
			Volume:    *quote.Volume[i],
		}
		if len(result.Indicators.AdjClose) > 0 && result.Indicators.AdjClose[0].AdjClose[i] != nil {
			bar.AdjClose = result.Indicators.AdjClose[0].AdjClose[i]
		}
		bars = append(bars, bar)
	}
	return bars, nil
}

// GetMetadata returns the chart metadata.
func (r *ChartResponse) GetMetadata() *ChartMeta {
	if len(r.Chart.Result) == 0 {
		return nil
	}
	return &r.Chart.Result[0].Meta
}

// GetQuotes extracts quotes from a quote response.
func (r *QuoteResponse) GetQuotes() []RawQuote {
	quotes := make([]RawQuote, 0, len(r.QuoteResponse.Result))
	for _, result := range r.QuoteResponse.Result {
		q := RawQuote{
			Symbol:                    result.Symbol,
			Currency:                  result.Currency,
			Exchange:                  result.Exchange,
			FullExchangeName:          result.FullExchangeName,
			ShortName:                 result.ShortName,
			LongName:                  result.LongName,
			QuoteType:                 result.QuoteType,
			MarketState:               result.MarketState,
			ExchangeTimezoneName:      result.ExchangeTimezoneName,
			ExchangeTimezoneShortName: result.ExchangeTimezoneShortName,
			GmtOffsetMilliseconds:     result.GmtOffsetMilliseconds,
		}
		if result.Bid != nil {
			q.Bid = result.Bid
		}
		if result.Ask != nil {
			q.Ask = result.Ask
		}
		if result.BidSize != nil {
			q.BidSize = result.BidSize
		}
		if result.AskSize != nil {
			q.AskSize = result.AskSize
		}
		if result.RegularMarketPrice != nil {
			q.RegularMarketPrice = result.RegularMarketPrice
		}
		if result.RegularMarketTime != nil {
			q.RegularMarketTime = result.RegularMarketTime
		}
		if result.RegularMarketChange != nil {
			q.RegularMarketChange = result.RegularMarketChange
		}
		if result.RegularMarketOpen != nil {
			q.RegularMarketOpen = result.RegularMarketOpen
		}
		if result.RegularMarketDayHigh != nil {
			q.RegularMarketDayHigh = result.RegularMarketDayHigh
		}
		if result.RegularMarketDayLow != nil {
			q.RegularMarketDayLow = result.RegularMarketDayLow
		}
		if result.RegularMarketVolume != nil {
			q.RegularMarketVolume = result.RegularMarketVolume
		}
		if result.RegularMarketChangePercent != nil {
			q.RegularMarketChangePercent = result.RegularMarketChangePercent
		}
		quotes = append(quotes, q)
	}
	return quotes
}

// GetFundamentals extracts fundamentals from a fundamentals response.
func (r *FundamentalsResponse) GetFundamentals() (*Fundamentals, error) {
	if len(r.QuoteSummary.Result) == 0 {
		return nil, ErrNoFundamentalsResults
	}
	result := r.QuoteSummary.Result[0]
	fundamentals := &Fundamentals{}
	if result.IncomeStatementHistoryQuarterly != nil {
		fundamentals.IncomeStatements = result.IncomeStatementHistoryQuarterly.IncomeStatementHistory
	}
	if result.BalanceSheetHistoryQuarterly != nil {
		fundamentals.BalanceSheets = result.BalanceSheetHistoryQuarterly.BalanceSheetHistory
	}
	if result.CashflowStatementHistoryQuarterly != nil {
		fundamentals.CashflowStatements = result.CashflowStatementHistoryQuarterly.CashflowStatementHistory
	}
	return fundamentals, nil
}

// ErrNoChartResults / ErrNoFundamentalsResults are sentinel errors used
// by the Get* accessors. They live in model/ because they describe
// model-level structural conditions (the Get* methods are defined on
// model types now).
var (
	ErrNoChartResults        = errString("model: chart response has no results")
	ErrNoFundamentalsResults = errString("model: fundamentals response has no results")
)

type errString string

func (e errString) Error() string { return string(e) }
