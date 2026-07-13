// yahoo_chart.go — raw Yahoo Finance chart/bars response structs.
// Originally lived in svc/yahoo/bars.go; promoted to model/ so external
// consumers (cmd, facade, model/normalize.go) can depend on the raw API
// shape without importing the Decode/Validate/Fetch behavior of svc/yahoo.
//
// Renaming note: the singular Yahoo bar is exposed as `ChartBar` here
// (rather than `Bar`) to avoid colliding with the existing `model.Bar`
// SDK-surface shape (post-decode, ScaledDecimal-stripped). svc/yahoo/bars.go
// re-exposes it as `yahoo.Bar` via a type alias so existing callers don't
// need to migrate.

package model

// ChartResponse represents the Yahoo Finance bars API response.
// (renamed from `BarsResponse` for consistency with the model/yahoo_*.go
// convention used by the plan in plans/ethereal-chasing-haven.md.)
type ChartResponse struct {
	Chart Chart `json:"chart"`
}

// Chart contains the chart data
type Chart struct {
	Result []ChartResult `json:"result"`
	Error  *string       `json:"error"`
}

// ChartResult contains the actual chart data for a symbol
type ChartResult struct {
	Meta       ChartMeta       `json:"meta"`
	Timestamp  []int64         `json:"timestamp"`
	Indicators ChartIndicators `json:"indicators"`
}

// ChartMeta contains metadata about the chart
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

// CurrentTradingPeriod contains trading period information
type CurrentTradingPeriod struct {
	Pre     *TradingPeriod `json:"pre"`
	Regular *TradingPeriod `json:"regular"`
	Post    *TradingPeriod `json:"post"`
}

// TradingPeriod represents a trading period
type TradingPeriod struct {
	Timezone  string `json:"timezone"`
	Start     int64  `json:"start"`
	End       int64  `json:"end"`
	GmtOffset int64  `json:"gmtoffset"`
}

// ChartIndicators contains the price and volume indicators
type ChartIndicators struct {
	Quote    []QuoteIndicator    `json:"quote"`
	AdjClose []AdjCloseIndicator `json:"adjclose"`
}

// QuoteIndicator contains OHLCV data
type QuoteIndicator struct {
	Open   []*float64 `json:"open"`
	High   []*float64 `json:"high"`
	Low    []*float64 `json:"low"`
	Close  []*float64 `json:"close"`
	Volume []*int64   `json:"volume"`
}

// AdjCloseIndicator contains adjusted close prices
type AdjCloseIndicator struct {
	AdjClose []*float64 `json:"adjclose"`
}

// ChartBar is a single raw Yahoo Finance OHLCV bar. Renamed from `Bar`
// to avoid colliding with the SDK-surface `model.Bar`; svc/yahoo/bars.go
// re-aliases this as `yahoo.Bar` for backward compatibility.
type ChartBar struct {
	Timestamp int64    `json:"timestamp"`
	Open      float64  `json:"open"`
	High      float64  `json:"high"`
	Low       float64  `json:"low"`
	Close     float64  `json:"close"`
	Volume    int64    `json:"volume"`
	AdjClose  *float64 `json:"adjclose,omitempty"`
}

// GetMetadata returns the chart metadata
func (r *ChartResponse) GetMetadata() *ChartMeta {
	if len(r.Chart.Result) == 0 {
		return nil
	}
	return &r.Chart.Result[0].Meta
}

// IsAdjusted returns true if adjusted close data is available
func (r *ChartResponse) IsAdjusted() bool {
	if len(r.Chart.Result) == 0 {
		return false
	}
	return len(r.Chart.Result[0].Indicators.AdjClose) > 0
}