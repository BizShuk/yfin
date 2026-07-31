# Go API Reference

`go doc` 與 exported declarations 是 signature 的 source of truth；本頁提供 surface map 與 semantic boundaries。

```bash
go doc github.com/bizshuk/yfin/facade
go doc -all github.com/bizshuk/yfin/facade.Client
go doc github.com/bizshuk/yfin/model
```

## Client Construction

```text
facade.NewClient()
facade.NewClientWithConfig(*httpx.Config)
```

一個 `Client` 同時服務 external SDK consumers 與 in-repo CLI。重用 Client 以共享 transport、cookie jar、rate limit、circuit breakers 與 MIC cache。

## Yahoo API Surface

Plain `float64` DTO methods：

- Daily / intraday / weekly / monthly bars
- Quote
- Quarterly fundamentals
- Company info
- Market data
- Annual income statement / balance sheet / cash flow
- News

Precision-preserving methods：

- `FetchDailyBarsNorm`
- `FetchQuoteNorm`
- `FetchFundamentalsNorm`
- `FetchMarketDataNorm`

`Fetch*Norm` 回傳 `model.Normalized*`。一般 application 使用 plain methods；只有需要 `ScaledDecimal` 時才選 normalized surface。

## Foundation Surface

```text
facade.FoundationCommands()
(*facade.Client).FetchFoundation(...)
```

這是 `batch` 與 downstream SDK 共用的 ordered command manifest / generic dispatch。Cache-aware consumer 使用 `NewCachedFoundationClient`，不自行複製 command-to-provider mapping。

## Scrape Surface

Explicit operations 包含：

- financials / balance sheet / cash flow
- key statistics
- analysis / analyst insights
- profile
- news
- all-fundamentals orchestration

完整 DTO methods（`ScrapeFinancialsData` / `ScrapeProfileData` / `ScrapeKeyStatisticsData` / `ScrapeAnalysisData`）與 plain snapshot methods 都掛在 `facade.Client`。Pure parser wrappers 也由 facade 暴露，讓 `cmd` 不 import `svc/scrape`。

Scrape 與 Yahoo API 是不同 surfaces；沒有 general automatic fallback。Standalone soak 的 `auto` strategy 只對有 equivalent contract 的 `fundamentals → financials` 使用 fallback。

## TWSE Surface

```text
facade.NewTwseClient()
(*facade.TwseClient).Dispatch(...)
facade.TwseRegistry
facade.TwseIsNoData(...)
```

`TwseClient` 是 facade-owned opaque handle。Caller 不應 import `svc/twse` 取得 concrete client。

## Cache Surface

| Constructor | Boundary |
| --- | --- |
| `NewCachedClient(rawRoot)` | Yahoo daily bars |
| `NewCachedTwseClient(rawRoot)` | TWSE endpoint requests |
| `NewCachedFoundationClient(rawRoot)` | Foundation commands |
| `NewRawCache(rawRoot)` | Legacy JSON reads |

Caller owns the selected raw root；yfin owns confinement、request identity、freshness 與 atomic persistence。

## DTO Ownership

Plain types canonically live in `model`：

- `Bar`, `BarBatch`
- `Quote`
- `MarketData`
- `CompanyInfo`
- `FundamentalsLine`, `FundamentalsSnapshot`
- `NewsItem`

`facade.*` aliases are compatibility surface。Field details：

```bash
go doc github.com/bizshuk/yfin/model.BarBatch
go doc github.com/bizshuk/yfin/model.MarketData
go doc github.com/bizshuk/yfin/model.NormalizedQuote
```

## Missing-Data Semantics

- `nil` pointer means provider did not report a nullable field.
- Empty chart series is valid no-data and preserves identifying metadata.
- Currency absence remains empty; yfin does not infer or invent it.
- Plain non-pointer numbers cannot distinguish missing from a reported zero.
- Times are UTC; daily trading date is derived from exchange timezone.

## Errors

`utils/httpx` exposes typed HTTP / transport errors and retry classifiers. Facade wraps errors with operation context, so callers should prefer `errors.Is` / `errors.As` where typed errors are available and treat message matching as compatibility fallback only.

Quarterly fundamentals may return a paid-subscription error. Scrape methods can be selected explicitly when their data contract fits; do not silently reinterpret all API failures as scrape requests.

## Configuration

`httpx.Config` owns shared transport policy. `config.Loader` maps YAML into this type and into scrape policy:

```go
loaded, err := config.NewLoader("config/effective.yaml").Load()
if err != nil {
	return err
}

client := facade.NewClientWithConfig(loaded.GetHTTPConfig())
if err := client.ConfigureScrape(loaded.GetScrapeConfig()); err != nil {
	return err
}
```

CLI consumers should use `cmd.CreateClientFromConfig` so flag overrides and command policy share one loaded snapshot.

## Compatibility Notes

- Session rotation is removed; `--sessions` is vestigial.
- OpenTelemetry tracer is currently no-op.
- Production CLI has no FX rate provider.
- Some normalized JSON metadata retains legacy `ampy.*` schema identifiers; there is no protobuf emission dependency.

## Verification

```bash
go test ./model ./facade ./utils/httpx
go test -tags=integration ./... -run '^$'
go vet ./...
```
