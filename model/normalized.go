// normalized.go — all `Normalized*` data structs and `Meta`. These are the
// ScaledDecimal-rich intermediate types emitted from `Normalize*` functions
// to the proto-emit pipeline and to JSON fixtures. Originally split across
// svc/norm/types.go (core types) + svc/norm/holders.go + svc/norm/insider.go;
// consolidated here for cross-layer reuse.

package model

import "time"

// NormalizedBar represents a normalized bar with UTC times and scaled decimals
type NormalizedBar struct {
	Start              time.Time     `json:"start"`
	End                time.Time     `json:"end"`
	Open               ScaledDecimal `json:"open"`
	High               ScaledDecimal `json:"high"`
	Low                ScaledDecimal `json:"low"`
	Close              ScaledDecimal `json:"close"`
	Volume             int64         `json:"volume"`
	Adjusted           bool          `json:"adjusted"`
	AdjustmentPolicyID string        `json:"adjustment_policy_id"`
	CurrencyCode       string        `json:"currency_code"`
	EventTime          time.Time     `json:"event_time"`
	IngestTime         time.Time     `json:"ingest_time"`
	AsOf               time.Time     `json:"as_of"`
}

// NormalizedBarBatch represents a batch of normalized bars
type NormalizedBarBatch struct {
	Security Security        `json:"security"`
	Bars     []NormalizedBar `json:"bars"`
	Meta     Meta            `json:"meta"`
}

// NormalizedQuote represents a normalized quote
type NormalizedQuote struct {
	Security            Security       `json:"security"`
	Type                string         `json:"type"`
	Bid                 *ScaledDecimal `json:"bid,omitempty"`
	BidSize             *int64         `json:"bid_size,omitempty"`
	Ask                 *ScaledDecimal `json:"ask,omitempty"`
	AskSize             *int64         `json:"ask_size,omitempty"`
	RegularMarketPrice  *ScaledDecimal `json:"regular_market_price,omitempty"`
	RegularMarketHigh   *ScaledDecimal `json:"regular_market_high,omitempty"`
	RegularMarketLow    *ScaledDecimal `json:"regular_market_low,omitempty"`
	RegularMarketVolume *int64         `json:"regular_market_volume,omitempty"`
	Venue               string         `json:"venue,omitempty"`
	CurrencyCode        string         `json:"currency_code"`
	EventTime           time.Time      `json:"event_time"`
	IngestTime          time.Time      `json:"ingest_time"`
	Meta                Meta           `json:"meta"`
}

// NormalizedFundamentalsLine represents a single fundamentals line item
type NormalizedFundamentalsLine struct {
	Key          string        `json:"key"`
	Value        ScaledDecimal `json:"value"`
	CurrencyCode string        `json:"currency_code"`
	PeriodStart  time.Time     `json:"period_start"`
	PeriodEnd    time.Time     `json:"period_end"`
}

// NormalizedFundamentalsSnapshot represents a normalized fundamentals snapshot
type NormalizedFundamentalsSnapshot struct {
	Security Security                     `json:"security"`
	Lines    []NormalizedFundamentalsLine `json:"lines"`
	Source   string                       `json:"source"`
	AsOf     time.Time                    `json:"as_of"`
	Meta     Meta                         `json:"meta"`
}

// NormalizedCompanyInfo represents normalized company information
type NormalizedCompanyInfo struct {
	Security         Security   `json:"security"`
	LongName         string     `json:"long_name"`
	ShortName        string     `json:"short_name"`
	Exchange         string     `json:"exchange"`
	FullExchangeName string     `json:"full_exchange_name"`
	Currency         string     `json:"currency"`
	InstrumentType   string     `json:"instrument_type"`
	FirstTradeDate   *time.Time `json:"first_trade_date,omitempty"`
	Timezone         string     `json:"timezone"`
	ExchangeTimezone string     `json:"exchange_timezone"`
	EventTime        time.Time  `json:"event_time"`
	IngestTime       time.Time  `json:"ingest_time"`
	Meta             Meta       `json:"meta"`
}

// NormalizedMarketData represents comprehensive market data
type NormalizedMarketData struct {
	Security             Security       `json:"security"`
	RegularMarketPrice   *ScaledDecimal `json:"regular_market_price,omitempty"`
	RegularMarketHigh    *ScaledDecimal `json:"regular_market_high,omitempty"`
	RegularMarketLow     *ScaledDecimal `json:"regular_market_low,omitempty"`
	RegularMarketVolume  *int64         `json:"regular_market_volume,omitempty"`
	FiftyTwoWeekHigh     *ScaledDecimal `json:"fifty_two_week_high,omitempty"`
	FiftyTwoWeekLow      *ScaledDecimal `json:"fifty_two_week_low,omitempty"`
	PreviousClose        *ScaledDecimal `json:"previous_close,omitempty"`
	ChartPreviousClose   *ScaledDecimal `json:"chart_previous_close,omitempty"`
	RegularMarketTime    *time.Time     `json:"regular_market_time,omitempty"`
	HasPrePostMarketData bool           `json:"has_pre_post_market_data"`
	CurrencyCode         string         `json:"currency_code"`
	EventTime            time.Time      `json:"event_time"`
	IngestTime           time.Time      `json:"ingest_time"`
	Meta                 Meta           `json:"meta"`
}

// NormalizedHolder represents a single holder entity
type NormalizedHolder struct {
	Organization string         `json:"organization"`
	PercentHeld  *ScaledDecimal `json:"percent_held,omitempty"`
	Position     *int64         `json:"position,omitempty"`
	Value        *int64         `json:"value,omitempty"`
}

// NormalizedHolders represents ownership data for a security
type NormalizedHolders struct {
	Security                Security           `json:"security"`
	InsidersPercentHeld     *ScaledDecimal     `json:"insiders_percent_held,omitempty"`
	InstitutionsPercentHeld *ScaledDecimal     `json:"institutions_percent_held,omitempty"`
	Institutional           []NormalizedHolder `json:"institutional"`
	MutualFund              []NormalizedHolder `json:"mutual_fund"`
	AsOf                    time.Time          `json:"as_of"`
	Meta                    Meta               `json:"meta"`
}

// NormalizedInsiderTxn represents a single insider transaction
type NormalizedInsiderTxn struct {
	FilerName string     `json:"filer_name"`
	Text      string     `json:"text"`
	Shares    *int64     `json:"shares,omitempty"`
	Value     *int64     `json:"value,omitempty"`
	Date      *time.Time `json:"date,omitempty"`
}

// NormalizedInsider represents insider transaction data for a security
type NormalizedInsider struct {
	Security     Security               `json:"security"`
	Transactions []NormalizedInsiderTxn `json:"transactions"`
	NetBuyShares *int64                 `json:"net_buy_shares,omitempty"`
	AsOf         time.Time              `json:"as_of"`
	Meta         Meta                   `json:"meta"`
}

// Meta contains metadata for normalized messages
type Meta struct {
	RunID         string `json:"run_id"`
	Source        string `json:"source"`
	Producer      string `json:"producer"`
	SchemaVersion string `json:"schema_version"`
}
