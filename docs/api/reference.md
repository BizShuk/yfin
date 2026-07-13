# API Reference

This document provides comprehensive documentation for the `yfin` Go library API — the public contract exposed via `github.com/bizshuk/yfin/facade`. All return types are plain Go structs with `float64` prices; the internal `*norm.ScaledDecimal` precision path is reserved for CLI / proto pipelines (see the bottom-of-file note).

## Import & Package Boundary

The public surface lives at:

```go
import "github.com/bizshuk/yfin/facade"
```

External consumers (e.g. `stock`, `data`) should import `facade` only — never the internal `svc/*` packages. Doing so keeps callers immune to internal refactors and avoids leaking the `ScaledDecimal` precision model.

## Client Creation

### `facade.NewClient()`

Creates a new Yahoo Finance client with default configuration — sensible timeouts, QPS rate limiting, retries, and circuit breaker.

```go
client := facade.NewClient()
```

### `facade.NewClientWithConfig(config *httpx.Config)`

Creates a new Yahoo Finance client with custom configuration.

```go
import (
    "time"

    "github.com/bizshuk/yfin/facade"
    "github.com/bizshuk/yfin/utils/httpx"
)

config := &httpx.Config{
    Timeout:     30 * time.Second,
    MaxAttempts: 3,
    QPS:         2.0,
}
client := facade.NewClientWithConfig(config)
```

> **Note (session rotation removed)** — earlier versions of this package exposed `NewClientWithSessionRotation()`. That constructor has been **deleted** in `yfin`; the single shared `http.Client` plus QPS limit + retries + circuit breaker is now the sole production path. Python `yfinance` still has session rotation; `yfin` does not.

## Historical Data Methods

### `FetchDailyBars()`

Fetch daily OHLCV (Open, High, Low, Close, Volume) data for a symbol.

```go
bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, adjusted, runID)
```

**Parameters**:

| Name        | Type        | Notes                                                  |
| ----------- | ----------- | ------------------------------------------------------ |
| `ctx`       | `context.Context` | Cancellation / timeouts.                         |
| `symbol`    | `string`    | Stock symbol (e.g. `"AAPL"`).                          |
| `start`     | `time.Time` | Inclusive.                                             |
| `end`       | `time.Time` | Exclusive.                                             |
| `adjusted`  | `bool`      | Apply split / dividend adjustments.                    |
| `runID`     | `string`    | Caller-supplied identifier for log / trace correlation.|

**Returns**: `*facade.BarBatch, error`.

**Data Structure**:

```go
type Bar struct {
    Date         string  `json:"date"`          // YYYY-MM-DD (UTC)
    Open         float64 `json:"open"`
    High         float64 `json:"high"`
    Low          float64 `json:"low"`
    Close        float64 `json:"close"`
    Volume       int64   `json:"volume"`
    Adjusted     bool    `json:"adjusted"`
    CurrencyCode string  `json:"currency_code"`
}

type BarBatch struct {
    Symbol string `json:"symbol"`
    MIC    string `json:"mic,omitempty"`
    Bars   []Bar  `json:"bars"`
}
```

**Example — loop bars**:

```go
for _, bar := range bars.Bars {
    fmt.Printf("%s  O=%.2f H=%.2f L=%.2f C=%.2f V=%d\n",
        bar.Date, bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
}
```

**Limitations**:

- Limited by Yahoo Finance's historical data availability.
- Some symbols may have incomplete data for certain date ranges.
- Intraday data not available through this method — use `FetchIntradayBars`.

### `FetchIntradayBars()`

Fetch intraday bars with configurable intervals.

```go
bars, err := client.FetchIntradayBars(ctx, "AAPL", start, end, interval, runID)
```

**Parameters**:

- `interval`: one of `"1m"`, `"5m"`, `"15m"`, `"30m"`, `"60m"`.

**Returns**: `*facade.BarBatch` (same shape as daily bars).

**Limitations**:

- May return HTTP 422 errors for some symbols.
- Data availability varies by symbol and exchange.
- Limited historical depth (typically 60 days for `1m`).

### `FetchWeeklyBars()`

Fetch weekly OHLCV data.

```go
bars, err := client.FetchWeeklyBars(ctx, "AAPL", start, end, adjusted, runID)
```

**Returns**: `*facade.BarBatch`.

### `FetchMonthlyBars()`

Fetch monthly OHLCV data.

```go
bars, err := client.FetchMonthlyBars(ctx, "AAPL", start, end, adjusted, runID)
```

**Returns**: `*facade.BarBatch`.

## Real-time Data Methods

### `FetchQuote()`

Get a current market quote snapshot for a symbol.

```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
```

**Returns**: `*facade.Quote, error`.

**Data Structure**:

```go
type Quote struct {
    Symbol    string    `json:"symbol"`
    Price     float64   `json:"price"`     // 0 if RegularMarketPrice was nil (closed market)
    Currency  string    `json:"currency"`
    EventTime time.Time `json:"event_time"`
}
```

`Quote.Price` is decoded once from the internal scaled-decimal — no manual scaling required.

**Example**:

```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if err != nil { return err }
fmt.Printf("%s: %.2f %s @ %s\n",
    quote.Symbol, quote.Price, quote.Currency, quote.EventTime.UTC().Format(time.RFC3339))
```

**Limitations**:

- No historical data.
- May be delayed (not strictly real-time).
- `Price` may be `0` when the regular market is closed (the underlying source returned a nil price).

### `FetchMarketData()`

Get comprehensive market data including 52-week ranges.

```go
marketData, err := client.FetchMarketData(ctx, "AAPL", runID)
```

**Returns**: `*facade.MarketData, error`.

**Data Structure**:

```go
type MarketData struct {
    Symbol              string    `json:"symbol"`
    MIC                 string    `json:"mic,omitempty"`
    RegularMarketPrice  *float64  `json:"regular_market_price,omitempty"`
    RegularMarketHigh   *float64  `json:"regular_market_high,omitempty"`
    RegularMarketLow    *float64  `json:"regular_market_low,omitempty"`
    RegularMarketVolume *int64    `json:"regular_market_volume,omitempty"`
    FiftyTwoWeekHigh    *float64  `json:"fifty_two_week_high,omitempty"`
    FiftyTwoWeekLow     *float64  `json:"fifty_two_week_low,omitempty"`
    PreviousClose       *float64  `json:"previous_close,omitempty"`
    CurrencyCode        string    `json:"currency_code,omitempty"`
    EventTime           time.Time `json:"event_time"`
}
```

`MarketData` keeps optional fields as `*float64` / `*int64` so callers can distinguish *missing* (nil) from *zero*. Use this when you need raw access to 52-week high/low and previous close — `FetchQuote` collapses these into a single `Price`.

## Company Information Methods

### `FetchCompanyInfo()`

Get basic company information from chart metadata.

```go
companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
```

**Returns**: `*facade.CompanyInfo, error`.

**Data Structure**:

```go
type CompanyInfo struct {
    Symbol           string `json:"symbol"`
    LongName         string `json:"long_name,omitempty"`
    ShortName        string `json:"short_name,omitempty"`
    Exchange         string `json:"exchange,omitempty"`
    FullExchangeName string `json:"full_exchange_name,omitempty"`
    Currency         string `json:"currency,omitempty"`
    InstrumentType   string `json:"instrument_type,omitempty"`
    Timezone         string `json:"timezone,omitempty"`
}
```

**Important Limitations**:

- Only basic security information is returned.
- Does NOT include: company address, executives, website, employees, business summary.
- For detailed company profiles, use alternative data sources or contribute to internal scrape pipelines.

## Fundamentals Methods

### `FetchFundamentalsQuarterly()`

Fetch quarterly financial fundamentals (requires paid Yahoo Finance subscription).

```go
fundamentals, err := client.FetchFundamentalsQuarterly(ctx, "AAPL", runID)
```

**Returns**: `*facade.FundamentalsSnapshot, error`.

**Data Structure**:

```go
type FundamentalsLine struct {
    Key          string    `json:"key"`              // e.g. "revenue", "eps_basic"
    Value        float64   `json:"value"`
    CurrencyCode string    `json:"currency_code,omitempty"`
    PeriodStart  time.Time `json:"period_start"`
    PeriodEnd    time.Time `json:"period_end"`
}

type FundamentalsSnapshot struct {
    Symbol string             `json:"symbol"`
    MIC    string             `json:"mic,omitempty"`
    Source string             `json:"source,omitempty"`
    AsOf   time.Time          `json:"as_of"`
    Lines  []FundamentalsLine `json:"lines,omitempty"`
}
```

**Important Limitations**:

- Requires Yahoo Finance paid subscription; a 401-class error surfaces as `fundamentals data requires Yahoo Finance paid subscription`.
- Limited to quarterly data only.
- A missing `Value` on a line is surfaced as `0` (Go `float64` cannot represent nil); use `PeriodStart` / `PeriodEnd` to detect genuinely empty rows.

## Scraping Methods (plain SDK output)

The scrape methods below use the same Yahoo HTML scraping engine that backs the CLI but expose plain SDK structs — no `ampy-proto` types leak through. They are available even when API endpoints require paid subscriptions.

### `ScrapeFinancials()`

Scrape the income statement.

```go
financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
```

**Returns**: `*facade.FundamentalsSnapshot`.

**Data Available**: revenue, expenses, net income, EPS, etc.

### `ScrapeBalanceSheet()`

Scrape balance sheet data.

```go
balanceSheet, err := client.ScrapeBalanceSheet(ctx, "AAPL", runID)
```

**Returns**: `*facade.FundamentalsSnapshot`.

### `ScrapeCashFlow()`

Scrape cash flow statement data.

```go
cashFlow, err := client.ScrapeCashFlow(ctx, "AAPL", runID)
```

**Returns**: `*facade.FundamentalsSnapshot`.

### `ScrapeKeyStatistics()`

Scrape key financial metrics and ratios.

```go
keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
```

**Returns**: `*facade.FundamentalsSnapshot`.

**Data Available**: P/E ratio, market cap, enterprise value, etc.

### `ScrapeAnalysis()`

Scrape analyst recommendations and price targets.

```go
analysis, err := client.ScrapeAnalysis(ctx, "AAPL", runID)
```

**Returns**: `*facade.FundamentalsSnapshot`.

### `ScrapeAnalystInsights()`

Scrape detailed analyst insights.

```go
insights, err := client.ScrapeAnalystInsights(ctx, "AAPL", runID)
```

**Returns**: `*facade.FundamentalsSnapshot`.

### `ScrapeNews()`

Scrape recent news articles and press releases.

```go
news, err := client.ScrapeNews(ctx, "AAPL", runID)
```

**Returns**: `[]facade.NewsItem, error`.

**Data Structure**:

```go
type NewsItem struct {
    Title       string    `json:"title,omitempty"`
    URL         string    `json:"url,omitempty"`
    Source      string    `json:"source,omitempty"`
    Summary     string    `json:"summary,omitempty"`
    PublishedAt time.Time `json:"published_at"`     // UTC
    Symbols     []string  `json:"symbols,omitempty"`
}
```

Note: `NewsItem` is a deliberately minimal view — proto-only fields (Id, SentimentScoreBp, IngestTime) are dropped; consumers that need them must drop down to the proto layer.

**Limitations**:

- May return empty results for some symbols.
- News availability varies by company and market activity.

### `ScrapeAllFundamentals()`

Scrape all available fundamentals data types in one call.

```go
snapshots, err := client.ScrapeAllFundamentals(ctx, "AAPL", runID)
```

**Returns**: `[]*facade.FundamentalsSnapshot`.

Includes financials, balance sheet, cash flow, key statistics, analysis, and analyst insights (one entry per type). Note: still a `[]*facade.FundamentalsSnapshot` slice — *not* a structured aggregate.

## Data Structure Conventions

### Field Naming (JSON tags)

- `Quote`, `BarBatch`, `MarketData`, `CompanyInfo`, `FundamentalsSnapshot`, `NewsItem` — JSON tags use **snake_case** (`regular_market_price`, `period_end`, `published_at`).

### Nullable vs Zero — Important

| Type                       | Convention                                                 |
| -------------------------- | ---------------------------------------------------------- |
| `Quote.Price`              | `0` when the source's price is missing (closed market).    |
| `MarketData.*Price`        | `*float64` — `nil` means "value not reported", not zero.   |
| `FundamentalsLine.Value`   | `0` means "missing or zero" (float64 cannot distinguish).  |

When you genuinely need to distinguish missing from zero, use `MarketData` (nullable pointers) over `Quote` (collapsed float).

### Facade vs Internal Precision

`facade.Client` returns plain structs with `float64` decimals — one conversion per field, all in `norm.FromScaledDecimal`. The internal pipeline still uses `*norm.ScaledDecimal` for the emit→ampy-proto path used by the CLI; that path is exposed via `FetchDailyBarsNorm`, `FetchQuoteNorm`, `FetchFundamentalsNorm`, and `FetchMarketDataNorm` on the same client. **External consumers should stick to the plain `Fetch*` / `Scrape*` methods** — the `*Norm` variants exist only for in-tree consumers (CLI's emit pipeline) that need wire-precision decimals.

## Error Handling

### Common Error Scenarios

1. **Empty news**

    - **Cause**: Yahoo Finance may not have news for certain symbols.
    - **Solution**: Handle gracefully — return an empty `[]NewsItem`, not an error.

2. **Paid-subscription walls**

    - **Cause**: `FetchFundamentalsQuarterly` requires Yahoo Finance paid access; 401 surfaces as `fundamentals data requires Yahoo Finance paid subscription`.
    - **Solution**: Fall back to `ScrapeFinancials` / `ScrapeKeyStatistics` for free-tier equivalent data.

3. **Closed market** (price `0`)

    - **Cause**: `FetchQuote` returns `Price = 0` when the regular market is closed.
    - **Solution**: Check `quote.EventTime`; treat `0` outside market hours as expected.

4. **HTTP 422 / rate limits**

    - **Cause**: Intraday endpoints and burst traffic can hit Yahoo's limits.
    - **Solution**: Use `context.WithTimeout` to bound retries; raise `httpx.Config.MaxAttempts` for noisy networks. There is **no session rotation** in `yfin` to fall back on.

5. **Missing company details**

    - **Cause**: `FetchCompanyInfo` only returns chart-derived security metadata.
    - **Solution**: Use alternative data sources for address / executives / web / employees.

## Best Practices

### Client Configuration

```go
// Default — fine for most workloads
client := facade.NewClient()

// Tight SLOs / noisy networks — tune httpx
client := facade.NewClientWithConfig(&httpx.Config{
    Timeout:     30 * time.Second,
    MaxAttempts: 5,
    QPS:         1.5,
})
```

### Rate Limiting

- Yahoo Finance has undocumented rate limits; the HTTP client enforces QPS and retries with exponential backoff.
- Use `context.WithTimeout` per call to cap wall-clock; do **not** introduce session rotation (the constructor no longer exists).

### Context Discipline

Always pass `ctx` through; the same client is safe to share across goroutines.

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

quote, err := client.FetchQuote(ctx, "AAPL", runID)
```

## Next Steps

- [Complete Examples](examples.md) — working code samples.
- [Migration Guide](../integrations/migration-guide.md) — moving from Python `yfinance` (or earlier versions of this SDK).
