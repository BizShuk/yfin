# Data Surface Comparison

選擇 method 的核心是 data contract，不是「哪個失敗就換另一個」。Yahoo API、Yahoo web scrape、TWSE 與 cache wrappers 各有不同 ownership 與 absence semantics。

## Surface Matrix

| Surface | Public entry | Best for | Important boundary |
| --- | --- | --- | --- |
| Yahoo chart / quote API | `facade.Client.Fetch*` | Bars、quote、company / market metadata | Provider JSON contract |
| Quarterly quoteSummary | `FetchFundamentalsQuarterly` | Quarterly statements | May require paid access |
| Annual timeseries | `FetchIncomeStatement` / `FetchBalanceSheet` / `FetchCashFlowStatement` | Latest annual statement | Separate free Yahoo timeseries surface |
| Yahoo web scrape | `facade.Client.Scrape*` | Profile、statistics、analysis、news、financial pages | Explicit caller choice; robots enforced |
| Foundation dispatch | `FoundationCommands` / `FetchFoundation` | Ordered 30-command parity | Generic artifact-oriented surface |
| TWSE open data | `facade.TwseClient.Dispatch` | Taiwan official exchange endpoints | Endpoint-specific flags / no-data |
| Read-through cache | `facade.NewCached*` | Reusing exact / freshness-qualified artifacts | yfin owns persistence policy |

## Plain vs Normalized

| Need | Choose |
| --- | --- |
| Simple application DTOs | Plain `Fetch*` |
| Decimal precision / nullable price fields | `Fetch*Norm` |
| Full scrape DTO details | `ScrapeFinancialsData` / `ScrapeProfileData` / `ScrapeKeyStatisticsData` or parser wrappers |
| Period line items | Plain `FundamentalsSnapshot` |

Plain values use `float64`; normalized values use `ScaledDecimal` where precision matters.

## Fallback Semantics

There is no general API-to-scrape fallback:

- Quote has no scrape equivalent in the supported contract.
- Daily bars have no scrape equivalent.
- Company/profile fields overlap only partially.
- Quarterly API fundamentals and web financial pages differ in shape and authorization.

Application code may explicitly choose a second surface when it can prove equivalence. Standalone soak restricts `auto` fallback to `fundamentals → financials`; every other endpoint / strategy pairing is rejected by validation.

## Selection Examples

### Price history

Use `FetchDailyBars` or `FetchDailyBarsNorm`. If the result is an empty batch, treat it as valid no-data and evaluate your market context; do not switch to scrape.

### Snapshot price

Use `FetchQuote` for a compact plain DTO or `FetchQuoteNorm` for nullable precision fields. `MarketData` provides a different, chart-metadata-oriented field set.

### Financial statements

- Quarterly statement requirement: `FetchFundamentalsQuarterly`.
- Latest annual statement: one of the annual timeseries methods.
- Web page metrics / projections: explicit `Scrape*`.

Do not merge their line items without defining period、currency、source 與 duplicate-key policy.

### Batch artifact collection

Use `FetchFoundation` / `NewCachedFoundationClient` or CLI `batch`. Do not duplicate the 30-command registry downstream.

### Taiwan exchange data

Use `TwseClient.Dispatch` and registry metadata. TWSE no-data is a typed condition distinguishable via `facade.TwseIsNoData`.

## Performance Considerations

- Reuse one `facade.Client`.
- Bound multi-symbol workers separately from per-host QPS.
- Scrape is usually larger / more parser-sensitive than JSON API, but fixed latency numbers are not a contract.
- Use facade cache wrappers when repeated exact requests are expected.
- Measure with active `yfin_*` HTTP metrics rather than documentation estimates.

## Data Quality Considerations

Before comparing two surfaces:

- align symbol / listing / MIC;
- align period / timestamp;
- require compatible currency when monetary values are compared;
- distinguish missing fields from zero;
- preserve source lineage.

See [Data Quality](../operations/data-quality.md) for absence semantics.

## Verification

```bash
go doc -all github.com/bizshuk/yfin/facade.Client
go test ./facade ./svc/yahoo ./svc/scrape ./svc/twse
```

Provider acceptance requires a separately recorded live integration run.
