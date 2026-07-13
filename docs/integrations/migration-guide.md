# Migration Guide: Python `yfinance` → Go `yfin`

This guide helps you migrate from Python's `yfinance` library to Go's `yfin` SDK. It includes feature comparisons, code examples, and best practices.

The Go SDK exposes its public surface through `github.com/bizshuk/yfin/facade`: a single `*facade.Client` with `Fetch*` (chart-API) and `Scrape*` (HTML-scraping) methods returning **plain Go structs with `float64` decimals** — no `*norm.ScaledDecimal` math at the call site.

## Table of Contents

1. [Feature Comparison](#feature-comparison)
2. [Key Differences](#key-differences)
3. [Migration Examples](#migration-examples)
4. [Data Structure Mapping](#data-structure-mapping)
5. [Error Handling Differences](#error-handling-differences)
6. [Performance Considerations](#performance-considerations)
7. [Best Practices](#best-practices)
8. [Migration Checklist](#migration-checklist)

## Feature Comparison

| Feature                  | Python yfinance                  | Go yfin                                                                | Notes                                                            |
| ------------------------ | -------------------------------- | ---------------------------------------------------------------------- | ---------------------------------------------------------------- |
| **Installation**         | `pip install yfinance`           | `go get github.com/bizshuk/yfin`                                       | Go requires Go 1.23+.                                            |
| **Import path**          | `import yfinance as yf`          | `import "github.com/bizshuk/yfin/facade"`                              | Facade is the only public surface.                               |
| **Client / ticker**      | `yf.Ticker("AAPL")`              | `facade.NewClient()`                                                  | One client is shared across symbols.                             |
| **Historical data**      | `ticker.history(...)`            | `client.FetchDailyBars(...)`                                          | Same shape — daily / weekly / monthly / intraday.                |
| **Real-time quote**      | `ticker.info`                    | `client.FetchQuote(...)`                                              | Collapses to a single `Quote.Price float64`.                     |
| **Company info**         | `ticker.info` (huge dict)        | `client.FetchCompanyInfo(...)`                                        | Limited fields; chart-metadata-derived only.                     |
| **Market data / 52wk**   | `ticker.info[...]`               | `client.FetchMarketData(...)`                                         | Nullable `*float64` distinguishes missing from zero.             |
| **Financials (paid)**    | `ticker.financials`              | `client.FetchFundamentalsQuarterly(...)`                              | Yahoo paid subscription required.                                |
| **Financials (free)**    | not available                    | `client.ScrapeFinancials(...)`                                        | HTML scrape, plain struct output.                                |
| **Balance sheet (free)** | not available                    | `client.ScrapeBalanceSheet(...)`                                      | HTML scrape, plain struct output.                                |
| **Cash flow (free)**     | not available                    | `client.ScrapeCashFlow(...)`                                          | HTML scrape, plain struct output.                                |
| **Key statistics**       | `ticker.info['marketCap']` etc.  | `client.ScrapeKeyStatistics(...)`                                     | HTML scrape.                                                     |
| **News**                 | `ticker.news`                    | `client.ScrapeNews(...)`                                              | Returns plain `[]facade.NewsItem`.                               |
| **Analyst data**         | `ticker.analysts` / `recommendations` | `client.ScrapeAnalysis(...)` + `ScrapeAnalystInsights(...)`        | Plain `FundamentalsSnapshot` output.                             |
| **Session rotation**     | built-in cookie rotation         | **not implemented** (removed)                                         | `NewClientWithSessionRotation` no longer exists; tune `httpx.Config` instead. |
| **Options data**         | `ticker.options`                 | ❌ Not supported                                                       | Not available.                                                  |
| **Institutional holders**| `ticker.institutional_holders`   | ❌ Not supported                                                       | Not available.                                                  |
| **Insider transactions** | `ticker.insider_transactions`    | ❌ Not supported                                                       | Not available.                                                  |
| **Data types**           | pandas `DataFrame`               | plain Go structs (`*facade.BarBatch`, `*facade.Quote`, ...)           | Strongly typed.                                                  |
| **Error handling**       | exceptions                       | explicit `error` returns                                              | Idiomatic Go.                                                   |
| **Concurrency**          | limited                          | built-in goroutines, semaphore helpers                                | Better concurrency story.                                        |
| **Performance**          | slow for large datasets          | faster for concurrent requests                                        | Go is generally faster.                                          |

## Key Differences

### 1. API Design Philosophy

**Python yfinance**:

```python
import yfinance as yf

ticker = yf.Ticker("AAPL")
info = ticker.info
history = ticker.history(period="1mo")
news = ticker.news
```

**Go yfin** (via the `facade` package):

```go
import (
    "context"
    "github.com/bizshuk/yfin/facade"
)

client := facade.NewClient()
ctx := context.Background()

quote, err := client.FetchQuote(ctx, "AAPL", "run-1")
bars, err  := client.FetchDailyBars(ctx, "AAPL", start, end, true, "run-1")
news, err  := client.ScrapeNews(ctx, "AAPL", "run-1")
```

### 2. Data Structure Differences

**Python yfinance** returns pandas DataFrames and dictionaries:

```python
history = ticker.history(period="1mo")
print(history.head())

info = ticker.info
print(info['longName'])
print(info['currentPrice'])
```

**Go yfin** returns plain structs:

```go
bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
for _, bar := range bars.Bars {
    fmt.Printf("%s  C=%.2f\n", bar.Date, bar.Close)
}

quote, err := client.FetchQuote(ctx, "AAPL", runID)
fmt.Printf("Price: %.2f %s\n", quote.Price, quote.Currency)
```

### 3. Error Handling

**Python yfinance** uses exceptions:

```python
try:
    ticker = yf.Ticker("AAPL")
    info = ticker.info
    print(info['longName'])
except Exception as e:
    print(f"Error: {e}")
```

**Go yfin** uses explicit `error` returns:

```go
client := facade.NewClient()
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Price: %.2f\n", quote.Price)
```

### 4. Session Rotation — Removed in `yfin`

Python `yfinance` rotates cookies/headers to dodge rate limits. `yfin` **does not** — the constructor was deleted (see `CLAUDE.md`'s "No Session Rotation" decision). The single shared `http.Client` plus QPS limit + retries + circuit breaker is the entire production story. To trade recovery against cost, tune `utils/httpx.Config` via `facade.NewClientWithConfig`:

```go
import (
    "time"

    "github.com/bizshuk/yfin/facade"
    "github.com/bizshuk/yfin/utils/httpx"
)

client := facade.NewClientWithConfig(&httpx.Config{
    Timeout:     30 * time.Second,
    MaxAttempts: 5,
    QPS:         1.5,
})
```

## Migration Examples

### Example 1: Basic Data Fetching

**Python yfinance**:

```python
import yfinance as yf

ticker = yf.Ticker("AAPL")
info = ticker.info
print(f"Company: {info['longName']}")
print(f"Price: ${info['currentPrice']}")

history = ticker.history(period="1mo")
print(f"Latest close: ${history['Close'].iloc[-1]:.2f}")

news = ticker.news
print(f"Latest news: {news[0]['title']}")
```

**Go yfin**:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bizshuk/yfin/facade"
)

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("example-%d", time.Now().Unix())

    companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
    if err != nil { log.Fatal(err) }
    fmt.Printf("Company: %s\n", companyInfo.LongName)

    quote, err := client.FetchQuote(ctx, "AAPL", runID)
    if err != nil { log.Fatal(err) }
    fmt.Printf("Price: %.2f %s\n", quote.Price, quote.Currency)

    end := time.Now()
    start := end.AddDate(0, 0, -30)
    bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
    if err != nil { log.Fatal(err) }
    if len(bars.Bars) > 0 {
        latest := bars.Bars[len(bars.Bars)-1]
        fmt.Printf("Latest close: %.2f\n", latest.Close)
    }

    news, err := client.ScrapeNews(ctx, "AAPL", runID)
    if err != nil {
        log.Printf("Warning: News failed: %v", err)
    } else if len(news) > 0 {
        fmt.Printf("Latest news: %s\n", news[0].Title)
    }
}
```

### Example 2: Multiple Symbols

**Python yfinance**:

```python
import yfinance as yf

symbols = ["AAPL", "MSFT", "GOOGL"]
tickers = yf.Tickers(" ".join(symbols))
for symbol in symbols:
    ticker = tickers.tickers[symbol]
    info = ticker.info
    print(f"{symbol}: ${info['currentPrice']}")
```

**Go yfin**:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/bizshuk/yfin/facade"
)

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    symbols := []string{"AAPL", "MSFT", "GOOGL"}

    var wg sync.WaitGroup
    results := make(chan string, len(symbols))

    for _, symbol := range symbols {
        wg.Add(1)
        go func(sym string) {
            defer wg.Done()
            runID := fmt.Sprintf("multi-%d", time.Now().Unix())
            quote, err := client.FetchQuote(ctx, sym, runID)
            if err != nil {
                results <- fmt.Sprintf("%s: Error - %v", sym, err)
                return
            }
            results <- fmt.Sprintf("%s: %.2f %s", sym, quote.Price, quote.Currency)
        }(symbol)
    }
    wg.Wait()
    close(results)
    for r := range results {
        fmt.Println(r)
    }
}
```

### Example 3: Financial Data

**Python yfinance**:

```python
import yfinance as yf

ticker = yf.Ticker("AAPL")
financials = ticker.financials
print("Revenue:", financials.loc['Total Revenue'].iloc[0])

balance_sheet = ticker.balance_sheet
print("Total Assets:", balance_sheet.loc['Total Assets'].iloc[0])

info = ticker.info
print("Market Cap:", info['marketCap'])
print("P/E Ratio:", info['trailingPE'])
```

**Go yfin** — *plain struct, no scaled-decimal math*:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bizshuk/yfin/facade"
)

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("financials-%d", time.Now().Unix())

    financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
    if err != nil { log.Fatal(err) }
    for _, line := range financials.Lines {
        if line.Key == "revenue" {
            fmt.Printf("Revenue: %.2f %s\n", line.Value, line.CurrencyCode)
            break
        }
    }

    balanceSheet, err := client.ScrapeBalanceSheet(ctx, "AAPL", runID)
    if err != nil {
        log.Printf("Warning: Balance sheet failed: %v", err)
    } else {
        for _, line := range balanceSheet.Lines {
            if line.Key == "total_assets" {
                fmt.Printf("Total Assets: %.2f %s\n", line.Value, line.CurrencyCode)
                break
            }
        }
    }

    keyStats, err := client.ScrapeKeyStatistics(ctx, "AAPL", runID)
    if err != nil {
        log.Printf("Warning: Key statistics failed: %v", err)
    } else {
        for _, line := range keyStats.Lines {
            switch line.Key {
            case "market_cap":
                fmt.Printf("Market Cap: %.2f %s\n", line.Value, line.CurrencyCode)
            case "pe_ratio":
                fmt.Printf("P/E Ratio: %.2f\n", line.Value)
            }
        }
    }
}
```

### Example 4: Data Processing (Moving Average)

**Python yfinance**:

```python
import yfinance as yf
import pandas as pd

ticker = yf.Ticker("AAPL")
history = ticker.history(period="1y")

history['MA20'] = history['Close'].rolling(window=20).mean()
history['MA50'] = history['Close'].rolling(window=50).mean()
history['Returns'] = history['Close'].pct_change()

print(history[['Close', 'MA20', 'MA50', 'Returns']].tail())
```

**Go yfin**:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bizshuk/yfin/facade"
)

type PriceData struct {
    Date   time.Time
    Close  float64
    MA20   float64
    MA50   float64
    Return float64
}

func movingAverage(prices []float64, window int) []float64 {
    if len(prices) < window {
        return make([]float64, len(prices))
    }
    ma := make([]float64, len(prices))
    for i := window - 1; i < len(prices); i++ {
        sum := 0.0
        for j := i - window + 1; j <= i; j++ {
            sum += prices[j]
        }
        ma[i] = sum / float64(window)
    }
    return ma
}

func returns(prices []float64) []float64 {
    out := make([]float64, len(prices))
    for i := 1; i < len(prices); i++ {
        out[i] = (prices[i] - prices[i-1]) / prices[i-1]
    }
    return out
}

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("analysis-%d", time.Now().Unix())

    end := time.Now()
    start := end.AddDate(-1, 0, 0) // 1 year ago
    bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
    if err != nil { log.Fatal(err) }

    prices := make([]float64, len(bars.Bars))
    dates  := make([]time.Time, len(bars.Bars))
    for i, bar := range bars.Bars {
        prices[i] = bar.Close                   // direct float64 — no scaling
        dates[i], _ = time.Parse("2006-01-02", bar.Date)
    }

    ma20 := movingAverage(prices, 20)
    ma50 := movingAverage(prices, 50)
    ret  := returns(prices)

    fmt.Println("Date\t\tClose\tMA20\tMA50\tReturn")
    fmt.Println("==================================================")
    for i := len(prices) - 5; i < len(prices); i++ {
        fmt.Printf("%s\t%.2f\t%.2f\t%.2f\t%.4f\n",
            dates[i].Format("2006-01-02"), prices[i], ma20[i], ma50[i], ret[i])
    }
}
```

## Data Structure Mapping

### Company Information

| Python yfinance            | Go yfin                          | Notes                                              |
| -------------------------- | -------------------------------- | -------------------------------------------------- |
| `info['longName']`         | `companyInfo.LongName`           | Company name.                                      |
| `info['shortName']`        | `companyInfo.ShortName`          | Short name.                                        |
| `info['exchange']`         | `companyInfo.Exchange`           | Exchange.                                          |
| `info['currency']`         | `companyInfo.Currency`           | Currency.                                          |
| `info['sector']`           | ❌ Not available                 | Use alternative sources.                           |
| `info['industry']`         | ❌ Not available                 | Use alternative sources.                           |
| `info['website']`          | ❌ Not available                 | Use alternative sources.                           |
| `info['address1']`         | ❌ Not available                 | Use alternative sources.                           |
| `info['city']`             | ❌ Not available                 | Use alternative sources.                           |
| `info['state']`            | ❌ Not available                 | Use alternative sources.                           |
| `info['zip']`              | ❌ Not available                 | Use alternative sources.                           |
| `info['country']`          | ❌ Not available                 | Use alternative sources.                           |
| `info['phone']`            | ❌ Not available                 | Use alternative sources.                           |
| `info['employees']`        | ❌ Not available                 | Use alternative sources.                           |
| `info['businessSummary']`  | ❌ Not available                 | Use alternative sources.                           |

### Quote Data

| Python yfinance            | Go yfin                                            | Notes                                                       |
| -------------------------- | -------------------------------------------------- | ----------------------------------------------------------- |
| `info['currentPrice']`     | `quote.Price`                                      | Direct `float64`. `0` means market closed or no price.      |
| `info['previousClose']`    | `marketData.PreviousClose`                         | `*float64` — `nil` if missing.                              |
| `info['dayHigh']`          | `marketData.RegularMarketHigh`                     | Nullable pointer.                                           |
| `info['dayLow']`           | `marketData.RegularMarketLow`                      | Nullable pointer.                                           |
| `info['volume']`           | `marketData.RegularMarketVolume`                   | `*int64`.                                                   |
| `info['marketCap']`        | `keyStats.Lines[Key="market_cap"].Value`           | From `ScrapeKeyStatistics`.                                 |
| `info['trailingPE']`       | `keyStats.Lines[Key="pe_ratio"].Value`             | From `ScrapeKeyStatistics`.                                 |

### Historical Data

| Python yfinance            | Go yfin                       | Notes                                                          |
| -------------------------- | ----------------------------- | -------------------------------------------------------------- |
| `history['Open']`          | `bar.Open`                    | `float64`.                                                     |
| `history['High']`          | `bar.High`                    | `float64`.                                                     |
| `history['Low']`           | `bar.Low`                     | `float64`.                                                     |
| `history['Close']`         | `bar.Close`                   | `float64`.                                                     |
| `history['Volume']`        | `bar.Volume`                  | `int64`.                                                       |
| `history.index[i]` (date)  | `bar.Date`                    | `string` `"YYYY-MM-DD"` (UTC).                                 |

## Error Handling Differences

### Python yfinance

```python
import yfinance as yf

try:
    ticker = yf.Ticker("INVALID_SYMBOL")
    info = ticker.info
    print(info['longName'])
except Exception as e:
    print(f"Error: {e}")
```

### Go yfin

```go
client := facade.NewClient()
quote, err := client.FetchQuote(ctx, "INVALID_SYMBOL", runID)
if err != nil {
    log.Printf("Error: %v", err)
    return
}
fmt.Printf("Price: %.2f\n", quote.Price)
```

`FetchFundamentalsQuarterly` additionally returns a wrapped error containing `"paid subscription"` when Yahoo Finance rejects the request — use `strings.Contains` to detect it and fall back to the `Scrape*` equivalents.

## Performance Considerations

### Python yfinance

- Slower for large datasets.
- Limited concurrency support.
- High memory usage with pandas DataFrames.

### Go yfin

- Faster for concurrent requests.
- Better memory efficiency.
- Goroutine support for concurrency.

**Example — Concurrent Processing**:

```go
func processSymbolsConcurrently(symbols []string) {
    client := facade.NewClient()
    ctx := context.Background()

    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10)

    for _, symbol := range symbols {
        wg.Add(1)
        go func(sym string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            runID := fmt.Sprintf("concurrent-%d", time.Now().Unix())
            quote, err := client.FetchQuote(ctx, sym, runID)
            if err != nil {
                log.Printf("Error fetching %s: %v", sym, err)
                return
            }
            log.Printf("%s: %.2f %s", quote.Symbol, quote.Price, quote.Currency)
        }(symbol)
    }
    wg.Wait()
}
```

## Best Practices

### 1. Construct once, share across goroutines

```go
// One client for the whole process.
client := facade.NewClient()
```

The shared client is goroutine-safe; building a fresh client per request just burns DNS / TCP handshakes.

### 2. Implement Proper Error Handling

```go
quote, err := client.FetchQuote(ctx, "AAPL", runID)
if err != nil {
    log.Printf("Error fetching quote: %v", err)
    return
}
fmt.Printf("Price: %.2f\n", quote.Price)
```

### 3. Read float64 fields directly — no scaling math

```go
// CORRECT — facade has already decoded.
price := quote.Price

// WRONG — `regularMarketPrice` doesn't exist on facade.Quote.
```

If you ever need the *internal* precision (e.g. piping into an internal normalized-struct consumer), use the matching `*Norm` variant: `FetchDailyBarsNorm`, `FetchQuoteNorm`, etc. — same call shape, returns `*model.NormalizedQuote` / `*model.NormalizedBarBatch` for in-tree consumers.

### 4. Use Context for Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

quote, err := client.FetchQuote(ctx, "AAPL", runID)
```

### 5. Implement Rate Limiting

```go
time.Sleep(1 * time.Second)            // simple fixed delay between symbols
semaphore := make(chan struct{}, 5)    // bounded concurrency
```

### 6. Validate Data Quality

```go
func validateQuote(quote *facade.Quote) error {
    if quote.Currency == "" {
        return fmt.Errorf("missing currency code")
    }
    return nil
}
```

`MarketData` is preferable when you need the *missing vs zero* distinction — its price/volume fields are `*float64` / `*int64`.

## Migration Checklist

- [ ] **Install Go yfin**: `go get github.com/bizshuk/yfin`.
- [ ] **Update imports**: Replace `import yfinance as yf` with `import "github.com/bizshuk/yfin/facade"`.
- [ ] **Use `facade.NewClient()`**: One client, shared goroutine-safe.
- [ ] **Update API calls**: Replace ticker property access with `client.Fetch*` / `client.Scrape*` methods.
- [ ] **Drop scaled-decimal math**: Read `float64` fields directly.
- [ ] **Implement error handling**: Replace `try/catch` with explicit `error` returns.
- [ ] **Add context**: Use `context.Context` for cancellation and timeouts.
- [ ] **Implement concurrency**: Use goroutines + semaphore for concurrent processing.
- [ ] **Add rate limiting**: Bound concurrency and pace requests.
- [ ] **Validate data**: Add data validation and quality checks.
- [ ] **Test thoroughly**: Test against real Yahoo Finance data.
- [ ] **Monitor**: Track latency, error rates, and 401 (paid-subscription) responses.

## Next Steps

- [API Reference](../api/reference.md) — complete API documentation.
- [Complete Examples](../api/examples.md) — working code samples.
