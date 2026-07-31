# Data Structures Guide

Canonical field definitions 位於 `model/*.go`；本頁只說明各 surface 的用途與 missing-data semantics。查看目前 struct：

```bash
go doc github.com/bizshuk/yfin/model.Bar
go doc github.com/bizshuk/yfin/model.NormalizedBar
go doc github.com/bizshuk/yfin/model.FundamentalsSnapshot
```

## Plain SDK DTOs

一般 consumer 優先使用 plain surface：

| Type | Purpose | Numeric representation |
| --- | --- | --- |
| `model.Bar` / `BarBatch` | Daily / interval OHLCV | `float64` prices |
| `model.Quote` | Compact quote snapshot | `float64` price |
| `model.MarketData` | Nullable current-market fields | `*float64` / `*int64` |
| `model.CompanyInfo` | Listing / company metadata | strings |
| `model.FundamentalsSnapshot` | Period-tagged line items | `float64` |
| `model.NewsItem` | Article metadata | strings / UTC time |

`facade.Bar`、`facade.Quote` 等是相同 types 的 compatibility aliases。資料 owner 是 `model`；operation owner 是 `facade.Client`。

## Precision-Preserving DTOs

需要 exact decimal transport 時使用：

- `NormalizedBarBatch`
- `NormalizedQuote`
- `NormalizedFundamentalsSnapshot`
- `NormalizedMarketData`
- `NormalizedCompanyInfo`

這些 types 由 `Fetch*Norm` methods 回傳，monetary fields 使用：

```go
type ScaledDecimal struct {
	Scaled int64
	Scale  int
}
```

轉為 display value：

```go
value := decimal.Float64()
// 或 model.FromScaledDecimal(decimal)
```

不要用 `Scaled / Scale`；`Scale` 是 decimal places，實際 divisor 是 `10^Scale`。

## Missing Data

不同 surface 的 absence semantics 不相同：

- Pointer fields：`nil` 表示 provider 未提供；不要改成 `0`。
- Plain `Quote.Price` / `FundamentalsLine.Value`：non-pointer `float64` 無法區分 missing 與 reported zero，caller 必須結合 source / time / surrounding fields 判讀。
- Empty chart series：structurally valid no-data result，保留 symbol、MIC、first-trade metadata，不是 transport error。
- Currency：provider 未提供時保留 empty string；不要猜 `USD`。
- News timestamp：缺值是 zero `time.Time`。

## Time Semantics

- Event / ingest boundaries 使用 UTC。
- Daily bar 的 `TradingDate` 依 exchange timezone 透過 `model.ToTradingDay` 推導，不以 UTC date 猜測交易日。
- `FundamentalsSnapshot.AsOf` 是已收錄 line items 中最新的 `PeriodEnd`。
- Invalid / missing quarterly period date 不以 current time 填補。

## Metadata

`model.Meta` 包含 run correlation 與 serialization identity：

- `RunID`：caller-provided correlation ID。
- `Source`：runtime identity `yfin`。
- `Producer`：process-local producer `local`。
- `SchemaVersion`：部分 normalized JSON 仍保留 legacy `ampy.*` logical schema identifier，作為 serialized fixture compatibility；這不代表存在 protobuf dependency 或 emission pipeline。

Provider lineage 則由 data surface 自身表示，例如 fundamentals `Source` 為 `yahoo-finance` 或 `yfin/scrape/...`。

## Raw / Provider Shapes

`model` 也包含 Yahoo raw envelopes、scrape DTOs 與 TWSE response rows，目的是維持 lowest-layer dependency direction。一般 application 不應繞過 facade 自行呼叫 provider transport。

## Data Conversion Rules

- Plain conversion：`facade.FromBarBatch`、`FromQuote`、`FromMarketData`、`FromFundamentalsSnapshot`。
- Scrape conversion：`model.Scrape*ToSnapshot` 與 `ScrapeNewsToItems`。
- FX conversion interfaces / mock types 只支援 explicit in-memory conversion logic；CLI production path 沒有 FX rate provider。

## Verification

```bash
go test ./model ./facade
go test -tags=integration ./... -run '^$'
```

Field-level assertions 應放在 tests，不複製到 Markdown。
