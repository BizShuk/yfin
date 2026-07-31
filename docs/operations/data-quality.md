# Data Quality and Absence Semantics

yfin 負責 decode、structural validation 與 normalization；business-specific quality gates 由 downstream application 擁有。本頁說明如何區分 transport failure、provider absence 與 application policy。

## Three Different Conditions

| Condition | Example | Handling |
| --- | --- | --- |
| Transport / provider failure | timeout、429、5xx、decode error | Return error；依 retry policy 處理 |
| Structurally valid no-data | chart 有 metadata 但沒有 observations | Return empty batch；保留 identity metadata |
| Missing field | nullable price、currency、timestamp 未提供 | Preserve `nil` / empty / zero-time semantics |

不要用 fixed defaults 把三者混在一起。特別是：

- 不猜 currency。
- 不補造 bars 或 quote prices。
- 不以 current time 代替缺少的 fundamentals period date。
- 不把 unavailable FX rate 報成 successful conversion。

## Plain vs Normalized Surface

Plain `model.Bar` / `Quote` / `FundamentalsLine` 使用 `float64`，適合一般 SDK consumer。Normalized DTO 使用 `ScaledDecimal` 與 pointer fields，能保留更多 absence / precision information。

需要判斷 provider 是否回報 market price 時，優先使用：

```go
quote, err := client.FetchQuoteNorm(ctx, symbol, runID)
if err != nil {
	return err
}
if quote.RegularMarketPrice == nil {
	// Provider did not report a price.
}
```

Plain `Quote.Price == 0` 可能代表 missing，也可能是真實 zero；不要僅靠該值建立 universal quality rule。

## Bars

Daily bar validation 在 observations 存在時檢查 symbol / OHLCV integrity。Empty series 仍可攜帶：

- symbol / MIC
- first trade date
- run metadata

Daily calendar date 必須透過 exchange timezone 推導。UTC storage 與 exchange-local trading date 是兩個不同概念。

Currency metadata 缺失時，bar currency 保持空白。若 downstream 擁有 official exchange context，可在自己的 boundary 補充，但必須保留 lineage。

## Quotes and Market Data

- `NormalizedQuote` 的 bid / ask / price / high / low 是 nullable。
- `MarketData` 的 price / volume fields 是 pointers。
- Event time 應為 UTC；missing provider timestamp 不應變造為 historical market event。
- MIC inference 只使用已知 exchange mapping；unknown exchange 可保留空 MIC。

## Fundamentals

Quarterly normalized fundamentals：

- `AsOf` 取已接受 line items 的最新 `PeriodEnd`。
- Missing / invalid period date 使該 statement 無效。
- QuoteSummary payload 未提供 currency 時，`CurrencyCode` 保持空白。
- `Source` 表示 provider surface，例如 `yahoo-finance` 或 `yfin/scrape/...`。

Revenue、income 或 assets 可能合法為 negative。Reasonableness thresholds 取決於 accounting domain、currency、scale 與 reporting period，不應硬編碼為全域 yfin 規則。

## Cross-Surface Comparison

比較兩個 surfaces 前先確認：

1. symbol / listing 是否相同。
2. reporting timestamp / period 是否可比。
3. currency 是否皆存在且相同。
4. adjustment policy 是否一致。
5. provider field definition 是否 equivalent。

Standalone soak correctness probes 只比較明確對應的 market cap、P/E、employees、sector、currency；missing value 會失敗，不使用 constants 補齊。Tolerance 是 probe policy，不是所有 SDK consumers 的 universal rule。

## Quality Gates Belong in Tests

可寫成 pass/fail 的 assertion 應放在 executable checks：

- `model/*_test.go`：normalization / absence semantics。
- `svc/*/*_test.go`：decode / parser fixtures。
- `facade/*_test.go`：public boundary conversion。
- `tests/data_correctness_test.go`：live provider comparisons（integration tag）。
- `cmd/soak/probes_test.go`：cross-source probe logic。

Deterministic suite：

```bash
go test ./...
go test -tags=integration ./... -run '^$'
```

Live suite：

```bash
go test -tags=integration ./...
```

Live failure 需記錄 upstream status 與 timestamp，不能直接視為 code regression。

## Operator Checklist

1. 保存 exact command、symbol、endpoint、run ID 與 UTC time。
2. 分辨 error、empty dataset、missing field。
3. 檢查 effective config、QPS、retry 與 timeout。
4. 對 parser drift 保存合法 fixture，新增 focused test。
5. 對 application-specific threshold 在 downstream 建立 owner 與 test。

## See Also

- [Data Structures](../api/data-structures.md)
- [Error Handling](error-handling.md)
- [Observability](observability.md)
- [Soak Testing](../cli/soak-testing.md)
