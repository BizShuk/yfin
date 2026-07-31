# Go API Examples

本頁提供最小 composition patterns。可編譯的完整程式由 [`facade/samples/`](../../facade/samples/) 擁有，避免文件與 API signature 分岔。

## Plain Daily Bars

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
 start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
 end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

 batch, err := client.FetchDailyBars(
  context.Background(),
  "AAPL",
  start,
  end,
  true,
  "example-bars",
 )
 if err != nil {
  log.Fatal(err)
 }
 for _, bar := range batch.Bars {
  fmt.Printf("%s %.2f %s\n", bar.Date, bar.Close, bar.CurrencyCode)
 }
}
```

Plain DTO fields 已是 `float64`，不需手動 scale arithmetic。

## Precision-Preserving Quote

```go
quote, err := client.FetchQuoteNorm(ctx, "AAPL", "example-quote")
if err != nil {
 return err
}
if quote.RegularMarketPrice != nil {
 fmt.Println(quote.RegularMarketPrice.Float64(), quote.CurrencyCode)
}
```

`nil` price 代表 provider 未提供；不要用假值補齊。

## Explicit Scrape Surface

```go
stats, err := client.ScrapeKeyStatistics(ctx, "AAPL", "example-scrape")
if err != nil {
 return err
}
fmt.Println(stats.Symbol, len(stats.Lines), stats.Source)
```

Scrape 是 caller 明確選擇的 surface。API methods 不會 general-purpose automatic fallback 到 HTML。

## Read-Through Cache

Downstream application 提供 raw root，由 yfin 管理 request identity 與 atomic JSON persistence：

```go
cached, err := facade.NewCachedClient("./runtime/raw")
if err != nil {
 return err
}

batch, err := cached.FetchDailyBars(
 ctx,
 "AAPL",
 start,
 end,
 true,
 "example-cache",
)
```

不要在 downstream 重複實作 Yahoo transport 或 cache write policy。

## Context and Cancellation

所有 network operation 都接受 `context.Context`：

```go
ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()

quote, err := client.FetchQuote(ctx, "MSFT", "example-timeout")
```

同一個 `facade.Client` 應跨 goroutines 重用；它擁有 shared connection pool、rate limiter、circuit breakers、cookie jar 與 MIC cache。

## Executable Samples

以下 commands 會呼叫 live providers：

```bash
go run ./facade/samples/historical_data
go run ./facade/samples/api_usage
go run ./facade/samples/print_all_data_types
go run ./facade/samples/print_data_contents
go run ./facade/samples/scrape_surface
```

Shell examples：

```bash
bash -n cmd/samples/*.sh
./cmd/samples/preview_aapl.sh
./cmd/samples/soak_smoke.sh --help
```

## Verify API Shape

```bash
go test ./facade ./model
go doc github.com/bizshuk/yfin/facade.Client
go doc github.com/bizshuk/yfin/model.Quote
```

Live sample success 不是 deterministic acceptance；CI / local unit tests 與 provider verification 應分開記錄。
