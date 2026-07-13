// yahoo_quote.go — raw Yahoo Finance quote response structs.
// Originally lived in svc/yahoo/quotes.go; promoted to model/ so external
// consumers (cmd, facade, model/normalize.go) can depend on the raw API
// shape without importing the Decode/Validate/Fetch behavior of svc/yahoo.
//
// Renaming note: the singular Yahoo quote is exposed as `RawQuote` here
// (rather than `Quote`) to avoid colliding with the existing `model.Quote`
// SDK-surface shape (post-decode). svc/yahoo/quotes.go re-exposes it as
// `yahoo.Quote` via a type alias so existing callers don't need to migrate.

package model

// QuoteResponse represents the Yahoo Finance quotes API response
type QuoteResponse struct {
	QuoteResponse QuoteResponseData `json:"quoteResponse"`
}

// QuoteResponseData contains the actual quote data
type QuoteResponseData struct {
	Result []QuoteResult `json:"result"`
	Error  *string       `json:"error"`
}

// QuoteResult contains quote data for a single symbol
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

// RawQuote is a single raw Yahoo Finance quote. Renamed from `Quote` to
// avoid colliding with the SDK-surface `model.Quote`; svc/yahoo/quotes.go
// re-aliases this as `yahoo.Quote` for backward compatibility.
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