# Migration Guide: Python `yfinance` to Go `yfin`

本指南只承諾 yfin 目前的 API。Python `yfinance` 的最新 surface 可能改變；遷移前應另外確認你現有 Python version 與實際使用欄位。

## Main Differences

| Concern | Python-style workflow | yfin |
| --- | --- | --- |
| Result shape | Often pandas / dict-like | Typed Go structs |
| Cancellation | Library-specific | `context.Context` |
| Errors | Exceptions | Explicit `error` |
| Precision | pandas / Python numeric | Plain `float64` or normalized `ScaledDecimal` |
| Concurrency | Application / library specific | Caller goroutines + yfin per-host policy |
| Cache | Application specific | Facade-owned read-through wrappers |
| Web extraction | May be implicit in library internals | Explicit `Scrape*` surface |

## Method Mapping

Use the data contract rather than the Python method name:

| Existing need | yfin entry |
| --- | --- |
| Historical daily bars | `Client.FetchDailyBars` |
| Intraday / weekly / monthly bars | Corresponding `Fetch*Bars` method |
| Current quote | `FetchQuote` / `FetchQuoteNorm` |
| Company/listing metadata | `FetchCompanyInfo` |
| Current market fields | `FetchMarketData` / `FetchMarketDataNorm` |
| Quarterly fundamentals | `FetchFundamentalsQuarterly` |
| Latest annual statements | `FetchIncomeStatement` / `FetchBalanceSheet` / `FetchCashFlowStatement` |
| Broad ticker command parity | `FetchFoundation` |
| Profile / statistics / analysis pages | Explicit `Scrape*` methods |
| News | `FetchNews` or explicit `ScrapeNews`, depending on desired source |

Quarterly fundamentals may require paid Yahoo access. Annual timeseries and scrape pages are separate contracts, not transparent replacements.

## Basic Migration

Python-style pseudocode:

```python
ticker = yf.Ticker("AAPL")
history = ticker.history(start="2026-01-01", end="2026-02-01")
```

Go:

```go
client := facade.NewClient()
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

batch, err := client.FetchDailyBars(
	ctx,
	"AAPL",
	start,
	end,
	true,
	"migration",
)
if err != nil {
	return err
}

for _, bar := range batch.Bars {
	fmt.Println(bar.Date, bar.Close, bar.CurrencyCode)
}
```

## DataFrame to Typed DTO

不要在 yfin layer 重建 pandas-style dynamic schema。先使用 typed DTO，再於 application boundary 轉成 table / database rows：

```go
type DailyClose struct {
	Symbol   string
	Date     string
	Close    float64
	Currency string
}

rows := make([]DailyClose, 0, len(batch.Bars))
for _, bar := range batch.Bars {
	rows = append(rows, DailyClose{
		Symbol:   batch.Symbol,
		Date:     bar.Date,
		Close:    bar.Close,
		Currency: bar.CurrencyCode,
	})
}
```

## Missing Data

Python `None` / `NaN` 不應一律改成 Go zero value：

- 使用 normalized pointer fields 判斷 missing price。
- Empty chart series 不是 transport error。
- Empty currency 表示 provider 未提供；不要補 `USD`。
- Plain non-pointer `float64` 無法區分 missing 與 zero；若需要區分，使用 normalized surface。

## Precision

Plain API：

```go
fmt.Println(quote.Price)
```

Normalized API：

```go
quote, err := client.FetchQuoteNorm(ctx, "AAPL", "migration")
if err != nil {
	return err
}
if quote.RegularMarketPrice != nil {
	fmt.Println(quote.RegularMarketPrice.Float64())
}
```

不要以 `Scaled / Scale` 計算；使用 `Float64` 或 `model.FromScaledDecimal`。

## Concurrency

Construct once, share across goroutines。`facade.Client` 是 concurrent workload 的 shared handler；每個 symbol 建立新 client 會拆散 connection pool、rate limiter、circuit breaker 與 MIC cache。

CLI multi-symbol `pull` / `quote` 已有 ordered bounded worker pool。SDK caller 若自行開 goroutines，仍受 `httpx` per-host QPS policy。

## Persistent Cache

若 Python workflow 已有本地 cache，遷移時改用：

```go
cached, err := facade.NewCachedClient(rawRoot)
```

Foundation / TWSE 有各自 wrapper。Consumer 可選 raw root，但不要自行複製 provider transport 或 persistence。

## Error Handling

```go
quote, err := client.FetchQuote(ctx, symbol, runID)
if err != nil {
	if errors.Is(err, context.DeadlineExceeded) {
		// application timeout policy
	}
	return err
}
```

HTTP typed errors 可用 `errors.Is` / `errors.As`。Scrape parser / robots errors 保持明確，不應被轉成 success with placeholder data。

## Migration Checklist

- [ ] 列出實際使用的 Python fields / periods / adjustments。
- [ ] 為每個 field 選 plain 或 normalized yfin surface。
- [ ] 定義 missing / zero / empty-series policy。
- [ ] 定義 currency 與 timezone policy。
- [ ] 重用一個 Client，設定 context timeout。
- [ ] 需要 cache 時使用 facade wrappers。
- [ ] 以 deterministic tests 固定 mapping。
- [ ] 另行記錄 live provider acceptance。

## Verification

```bash
go test ./...
go test -tags=integration ./... -run '^$'
go vet ./...
```

完整 current API 以 `go doc -all github.com/bizshuk/yfin/facade.Client` 為準。
