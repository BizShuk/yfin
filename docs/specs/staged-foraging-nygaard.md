# Plan: `cmd → facade → svc` 單一路徑架構重構

## Context

目前 `cmd/` 17 個檔案直接 import `svc/*`：
- `svc/yahoo`（3 處）、`svc/norm`（4）、`svc/emit`（4）、`svc/scrape`（8）、`svc/twse`（1）

設計初衷是 cmd 的 emit→proto pipeline 需要 `ScaledDecimal` 精度（plain SDK 已 float64-rounded），所以刻意繞過 facade。但這個繞路讓 cmd/ 同時持有：
1. `cmd/client.go` 的 `*CliClient`（包裝 `*yahoo.Client`）
2. `cmd/fetch.go` 的 4 個 exported helpers（`FetchDailyBarsNorm` 等，回傳 `*norm.Normalized*`）
3. `cmd/scrape/` 的 `buildScrapeURL` + `createScrapeClient`（直接 `svc/scrape.NewClient`）
4. `cmd/fundamentals/` 的 `scrapeNewClient` + `buildScrapeURL` 重複副本
5. `cmd/twse/` 的 23 個 `svc/twse.Fetch*` endpoints
6. `cmd/dispatch/` 的 30+ 個 entry 透過 `*yahoo.Client` 直接呼叫

使用者決策：強制執行 `cmd → facade(handler) → svc` 單一路徑，cmd 不再直接 import svc。

## 目標架構

```
cmd/ ─┬─ admin/     ─┐
      ├─ market/    ─┤
      ├─ fundamentals/ ─┼─→ facade ─→ svc/yahoo
      ├─ scrape/    ─┤          ├→ svc/norm
      ├─ dispatch/  ─┤          ├→ svc/scrape
      └─ twse/      ─┘          └→ svc/twse
```

facade 成為單一 handler，所有 cmd 對下層的呼叫都經過 facade。

## facade 需要新增的能力

| 既有 | 新增 |
|------|------|
| `FetchDailyBars` → `*BarBatch` | `FetchDailyBarsNorm` → `*norm.NormalizedBarBatch` |
| `FetchQuote` → `*Quote` | `FetchQuoteNorm` → `*norm.NormalizedQuote` |
| `FetchFundamentalsQuarterly` → `*FundamentalsSnapshot` | `FetchFundamentalsNorm` → `*norm.NormalizedFundamentalsSnapshot` |
| `FetchMarketData` → `*MarketData` | `FetchMarketDataNorm` → `*norm.NormalizedMarketData` |
| `ScrapeFinancials` 等 8 個回傳 `*FundamentalsSnapshot` | `ScrapeFetch` → `([]byte, *scrape.FetchMeta, error)`；`ParseComprehensiveFinancials` 等 7 個 parser wrapper |
| 沒有 | `BuildScrapeURL(ticker, endpoint)` |
| 沒有 | `TwseDispatch(ctx, endpoint, date, opts)` 統一入口 |
| 沒有 | `FetchInfo` / `FetchActions` / `FetchMetadata` / `FetchHolders` / `FetchInsider` / `FetchUpgrades` / `FetchCalendar` / `FetchEarningsDates` / `FetchSecFilings` / `FetchESG` / `FetchRecommendationTrend` / `FetchOptions` / `FetchISIN` |

觀察 8764 提醒：facade 既有 struct 都遵守 nullable/zero 區分（`*float64` / `*int64`），新增方法需維持這個 convention。

## 階段拆分

按依賴順序，5 個階段：

### Phase A — Norm-returning helpers + 拆 `cmd/fetch.go`
新增 4 個 facade method（`FetchDailyBarsNorm` / `FetchQuoteNorm` / `FetchFundamentalsNorm` / `FetchMarketDataNorm`），刪除 `cmd/fetch.go`。
- `cmd/market/pull.go` 改用 `client.FetchDailyBarsNorm(ctx, symbol, start, end, adjusted, runID)`
- `cmd/market/quote.go` 改用 `client.FetchQuoteNorm(ctx, symbol, runID)`
- `cmd/fundamentals/fundamentals_run.go` 改用 `client.FetchFundamentalsNorm(ctx, symbol, runID)`
- 風險：低；純內部替換，外部介面零變動。

### Phase B — 拆 `cmd/client.go` 的 CliClient
刪除 `cmd.CreateClient()` 與 `*CliClient`，全面改用 `facade.NewClient()` / `facade.NewClientWithConfig()`。
- 影響：`cmd/market/{pull,quote}.go`、`cmd/fundamentals/fundamentals_run.go`、`cmd/scrape/scrape_run.go`
- 風險：中。需把 `cmd.Global.QPS / RetryMax / Timeout` 的 CLI flag override 邏輯從 `CreateClient` 搬到 facade 的 constructor overload 或保留 facade 的預設行為。

### Phase C — scrape 走 facade
新增 facade 的 scrape helper API：
- `Client.ScrapeFetch(ctx, ticker, endpoint) ([]byte, *scrape.FetchMeta, error)` — 包裝 `scrape.Client.Fetch`
- 7 個 `Parse*` 函式（package-level 或 method）包裝 `svc/scrape.ParseComprehensive*`
- `BuildScrapeURL(ticker, endpoint) string`

影響：
- `cmd/scrape/scrape_run.go`：改用 `client.ScrapeFetch` + `facade.ParseComprehensive*`
- `cmd/fundamentals/{stats_helpers.go, profile.go, stats.go}`：移除 `scrapeNewClient` + `buildScrapeURL` 重複副本，改用 facade
- 風險：中。scrape 模式多，每個 endpoint case 都要導線。

### Phase D — TWSE 走 facade
新增 `Client.TwseDispatch(ctx, endpoint, date, opts) (any, error)` 統一入口，包裝 `svc/twse.Registry` 與 23 個 fetcher。
- `cmd/twse/twse.go` 的 `twseNameToFetcher` map 與 `runTwseEndpoint` 簡化為呼叫 facade
- `cmd/twse/client.go` 的 `buildTWSEClient` 也搬進 facade（或保留為 facade 的 private helper）
- 風險：低。23 個 endpoint 都是純 fetch + JSON encode，沒狀態。

### Phase E — dispatch 走 facade
`FetchContext` 移除 `*yahoo.Client` 欄位，只留 `*facade.Client`。
- 為 13 個 Python-yfinance 風格 commands（`info` / `actions` / `metadata` / `holders` / `insider` / `upgrades` / `calendar` / `earnings-dates` / `sec-filings` / `sustainability` / `recommendations` / `options` / `isin`）補上 facade method
- 影響：`cmd/dispatch/dispatch.go` 的 30+ entry 從 `fc.Y.FetchXxx` 改為 `fc.Root.FetchXxx`
- 風險：高。這 13 個 method 大多 Python yfinance 特有，外部不會用；考慮是否要全部加進 facade，或在 facade 內以 `Dispatch(ctx, name, args...)` 動態分派。

## Critical files

### 修改
- `facade/client.go` — 新增 Phase A/E 的 method
- `facade/scrape.go`（新檔）— Phase C 的 scrape helper
- `facade/twse.go`（新檔）— Phase D 的 twse dispatch
- `cmd/client.go` — 刪除大部分內容
- `cmd/fetch.go` — 完全刪除
- `cmd/scrape/scrape_run.go` — 改用 facade
- `cmd/scrape/format*.go` — 不變（純 formatter，與 svc 解耦）
- `cmd/market/{pull,quote}.go` — 改用 facade
- `cmd/fundamentals/*` — 移除重複的 scrapeNewClient / buildScrapeURL
- `cmd/dispatch/dispatch.go` — 簡化 FetchContext
- `cmd/twse/{twse.go, client.go}` — 改用 facade

### 不變
- `cmd/scrape/format*.go`（DTO → stdout formatter 是純 UI，與 svc 解耦）
- `cmd/{exitcodes,global,build,root}.go`
- `cmd/market/{market,client_json,pull_test}.go`
- `cmd/dispatch/{batch,batch_test,dispatch_test}.go`

## Verification

1. `grep -r '"github.com/bizshuk/yfin/svc' cmd/ --include='*.go'` 必須回空（除 `cmd/soak/` 既有 facade 介接外）。
2. `go build ./cmd/...` 通過
3. `go vet ./cmd/...` 通過
4. `go test ./cmd/...` 全綠
5. `./yfin --help` 10 個 subcommand 全部仍正確註冊
6. facade method 數：既有 16 → 預期增至 16 + 4 (Phase A) + 1 ScrapeFetch + 7 Parse + 1 BuildScrapeURL + 1 TwseDispatch + 13 dispatch = 43

## Out of scope（建議保持現狀）

- **bus publishing 仍留在 cmd**：`emit.EmitBarBatch` + `bus.PublishBars` 是 transport-level 邏輯，不屬於 data fetching 的 handler 職責。cmd/market/{pull,quote}.go 的 `handleBusPublishing` / `handleLocalExport` 維持不動。
- **DTO formatter 仍留在 cmd**：format.go / format_comprehensive.go 是 UI 邏輯（DTO → stdout），與 svc 解耦。
- **client_json.go 的 writeJSONFile**：cmd 的本地匯出，與 svc 無關。

## 風險

| 風險 | 緩解 |
|------|------|
| Phase E 新增 13 個 facade method 但外部不會用 | 改用 `facade.Dispatch(ctx, command, args...)` 動態分派單一入口；或把 13 個 method 標 `// yfinance-parity` 並放在 `facade_yfinance.go` 子檔，視覺上區隔 |
| Phase B 失去 `cmd.Global.{QPS,RetryMax,Timeout}` 的 CLI flag override | facade constructor 多一個 `WithOverrides(...)` option pattern，或 cmd 在呼叫 `facade.NewClient()` 前透過 `httpx.Config` 注入 |
| Phase C scrape 模式多，scrape_run.go 重寫量大 | scrape_run.go 仍由 cmd 擁有（控制 cobra 流程），只是把 `client.Fetch` + `scrape.Parse*` 換成 `facade.ScrapeFetch` + `facade.Parse*`；runner 邏輯不變 |
| 大規模改動一週 commit 數過大 | 5 個 Phase 各自可獨立 build/test，逐步 merge |
