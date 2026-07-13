# 移除四個 ampy* 依賴 + scrape 路徑改造

## Context

`yfin` 目前依賴四個 `AmpyFin/*` 套件(ampy-config / ampy-observability / ampy-bus / ampy-proto),設計目的是作為 Ampy 生態系的「上游資料生產者」——把 Yahoo 資料轉成 ampy-proto 訊息、經 ampy-bus envelope 發布到 NATS、用 ampy-observability 對接 inf 觀測後端、用 ampy-config 統一配置。

但用戶實際只寫本地檔:`yfin pull` / `yfin quote` / `yfin scrape` 的本地 JSON 匯出。已驗證 `handleLocalExport`(`cmd/market/pull.go:385`)走 `json.Encoder` 直序列化 `model.NormalizedBarBatch`,**完全不碰 emit/proto/bus**。`--publish` 從未使用。因此四個依賴全是 dead weight。

**目標**:移除四個 ampy* 依賴,scrape 路徑改走「scrape DTO 直通 model」(跳過 proto 中介),恢復 yfin 為獨立 Yahoo Finance Go client。

**已定決策**:
1. `cmd/scrape --preview-proto` 模式直接刪除(連旗標、函式、proto-typed printer),保留 `--preview-json` / `--preview-news` / `--check`。
2. scrape DTO → model 直轉換器放 `model/` 新檔(符合 model 最底層 + facade 包裝的分層慣例)。

**執行原則**:依風險由低到高分 4 phase,每 phase 結束 `go build ./...` + `go test ./...` 通過才進下一步。

---

## Phase 1 — 移除 ampy-config(最低風險)

`config/types/loader.go` 的 `ampyconfig.NewLoader(path).Load()` 只是把 YAML 讀成 `map[string]interface{}`,無 secret 解析 / 無多檔 merge / 無 hot-reload(已查 module source 確認)。secret 的 `ref` redact 與 `${VAR:-default}` 插值都是本地實作,不依賴 ampy-config。

### 改動
- `config/types/loader.go`:移除 `ampyconfig` import(L17);新增 private helper `loadYAMLFile(path) (map[string]interface{}, error)`(用 `os.ReadFile` + `yaml.Unmarshal`);替換 `Load()`(L37-38)與 `GetEffectiveConfig()`(L216-217)兩處 `ampyLoader.Load()` 呼叫。`interpolateEnvVars` / `mapToConfig` / `redactSecrets` / `validate` 等本地函式不動。
- `go.mod`:移除 `github.com/AmpyFin/ampy-config/go/ampyconfig v1.1.4`;`go mod tidy`。

### 驗證
```bash
go build ./... && go test ./config/...
./yfin admin config-effective   # 仍印 config
```

公開 API(`NewLoader`/`Load`/`GetEffectiveConfig`)不變,下游 7 個間接呼叫點(cmd/admin、cmd/client、cmd/fundamentals、cmd/market、cmd/scrape、cmd/soak)透明。

---

## Phase 2 — 移除 ampy-observability(中風險,內部重寫)

`utils/obsv/obsv.go` 是唯一 import `ampyobs` 的檔案。`Tracer()` 早已退化成 `noop`(註解坦承 ampy-obs 不暴露 tracer)。`utils/obsv/metrics.go` 是純本地 Prometheus(`prometheus/client_golang`),**0 個 ampyobs 呼叫,不動**。全倉庫 `obsv.` 呼叫點 35 處(httpx 23 處 + 4 個 cmd bootstrap),**只要 obsv.go 公開 API 簽名不變就透明**。

### 改動
- `utils/obsv/obsv.go`:重寫內部實作——
  - `ampyobs.Init` → 用 `slog.NewTextHandler(os.Stderr, ...)` 建 logger(從 `config.LogLevel` 解析 level),`otel.SetTracerProvider` 接 OTel SDK(已是 indirect dep)。OTLP exporter 構造失敗時 fallback `noop.NewTracerProvider()`,不擋啟動。
  - `ampyobs.L()` → 回傳自建 `*slog.Logger`;未初始化時 fallback `slog.Default()`(行為不變)。
  - `ampyobs.StartSpan` → `Tracer().Start(ctx, name, opts...)`(改用 OTel tracer,但 `globalObsv==nil` 時仍 noop)。
  - `ampyobs.Shutdown` → `shutdownMetrics` + tracer provider 的 `Shutdown(ctx)`。
  - `Observability` struct:移除 `ampyConfig` 欄位,改持有 `*slog.Logger` + `trace.TracerProvider`。
- 公開 API 簽名維持不變:`Init` / `Shutdown` / `Reset` / `Logger` / `Tracer` / `StartSpan` / `StartRunSpan` / `StartIngestFetchSpan` / `UpdateIngestFetchSpan` / `StartIngestDecodeSpan` / `StartIngestNormalizeSpan` / `StartEmitProtoSpan` / `StartPublishBusSpan` / `StartFXRatesSpan` / `RecordSpanError` / `LogWithTrace` / `CommonLogAttrs` + 7 個 span name 常數。`StartRunSpan` 等 7 個 helper 完全不動(已透過 `StartSpan` 走 OTel)。
- `go.mod`:移除 `github.com/AmpyFin/ampy-observability/go/ampyobs v0.0.3`;`go mod tidy`。
- `gosdk v1.1.0` / `ampy-bus` / `ampy-proto` 都不依賴 ampy-observability,無衝突。

### 驗證
```bash
go build ./... && go test ./utils/obsv/...
./yfin quote AAPL            # stderr 看到 slog text log
curl :9090/metrics           # Prometheus 仍正常
```

---

## Phase 3 — 移除 ampy-bus(中風險,整包刪 + 5 個 cmd 清 --publish)

`utils/bus/` 整包是 ampy-bus 的 NATS 封裝(envelope + chunking + topic + retry + circuit breaker)。`BusPublisher` 的 `proto.Marshal` 是唯一直接 marshal proto 的地方。`--publish` 是 `pull` / `quote` / `soak` 三命令的一級功能,用戶不用,整條路徑刪除。

### 刪除
- `utils/bus/` 整包(8 .go + 4 _test.go):`bus.go` / `types.go` / `topic.go` / `chunking.go` / `envelope.go` / `preview.go` / `publisher.go` / `retry.go`。
- `config/types/bus.go` 整檔。
- `config/effective.yaml` 的 `bus:` 區塊 + `nats_password` / `kafka_sasl_password` secret 條目;3 個 `config/example.*.yaml` 同樣清理。
- `tests/` 下若有 bus 相關測試一併清。

### 改寫
- `cmd/client.go`:刪 `CreateBusConfig` 整函式(L81-161)+ `utils/bus` import(L17)+ 標頭註解提及處。
- `cmd/market/pull.go`:刪 `handleBusPublishing`(L348-383)+ `estimateBarBatchSize`(L422-427)+ bus 建立區塊(L157-168)+ `processSymbol` 簽名移除 bus 參數(L266)+ `printBarsPreview` 簡化移除 env/topicPrefix 參數(L280)+ `pullConfig` 移除 `Publish`/`Preview`/`Env`/`TopicPrefix`/`DryRunPublish` 欄位與對應 5 個 cobra flag(L33-40, L66-72)+ L350 的 `emit.EmitBarBatch` 呼叫(只服務 --publish,隨分支刪)。`handleLocalExport`(L385)**完全不動**。
- `cmd/market/quote.go`:同 pull.go 模式——刪 `handleQuoteBusPublishing`(L181-216)+ `estimateQuoteSize`(L243-247)+ bus 建立區塊(L85-96)+ `processQuote` 簽名簡化(L132)+ `quoteConfig` 移除 4 欄位與 flag(L25-30, L49-52)+ L183 的 `emit.EmitQuote` 呼叫。
- `cmd/soak/orchestrator.go`:刪 `bus *bus.Bus` 欄位(L47)+ bus 建立區塊(L125-137)+ `o.bus.Close`(L427-429)+ `SoakConfig` 移除 `Publish`/`Preview`/`Env`/`TopicPrefix`(L33-40)+ `NewWorker` 呼叫移除 bus 參數(L250)。
- `cmd/soak/worker.go`:刪 `bus *bus.Bus` 欄位(L24)+ `NewWorker` 的 bus 參數(L39,43,44)——此欄位從未被使用是 dead state。
- `cmd/soak/main.go`:刪 `--preview` / `--publish` / `--env` / `--topic-prefix` 4 flag(L60-63)。
- `cmd/exitcodes.go`:刪 `ExitPublishError = 4`(從未被 `os.Exit` 呼叫,是預留碼)。
- `cmd/exitcodes_test.go`:刪 L18 對 `ExitPublishError` 的斷言。
- `config/types/config.go`:刪 `Bus BusConfig` 欄位(L19)。
- `config/types/adapters.go`:刪 `GetBusConfig`(L41-44)+ 註解;`GetHTTPConfig`/`GetFXConfig`/`GetScrapeConfig` 保留(純 struct getter)。
- `config/types/loader.go`:刪 bus 驗證區塊(L175-178, L190-195)(Phase 1 若已動則確認)。
- `config/types/loader_test.go`:刪 bus 測試資料 + `TestGetBusConfig`(L236-262)。
- `go.mod`:移除 `github.com/AmpyFin/ampy-bus v1.2.0`;`go mod tidy` 清 indirect(`nats-io/*` / `castai/promwrite` / `spf13/viper` 及其鏈 / `golang/snappy` / `grpc-gateway` 等)。

### 驗證
```bash
go build ./... && go test ./...
./yfin pull AAPL --start 2024-01-01 --end 2024-01-31 --preview --config config/effective.yaml  # 本地 JSON 仍產出
./yfin quote AAPL --preview --config config/effective.yaml
./yfin --help pull            # 不再列 --publish / --preview
grep -rn "utils/bus\|bus\.Bus\|ExitPublishError" --include='*.go' . | grep -v _test  # 0 hits
```

> 註:此 phase 後 `emit.EmitBarBatch` / `EmitQuote` 在 pull/quote 的呼叫點已刪,但 `svc/emit` 套件仍存在(被 facade.Scrape* 與 cmd/scrape 用),Phase 4 才整包刪。`google.golang.org/protobuf` 與 `ampy-proto` 此時仍保留。

---

## Phase 4 — 移除 ampy-proto + scrape 路徑改造(最高風險)

`svc/emit/` 整層(12 非測試檔 + 5 測試檔,4529 行)把 `model.Normalized*` / scrape DTO map 成 ampy-proto 訊息。移除後 `facade.Scrape*` 改走 `scrape.Parse*DTO → model.<新 converter>(dto) → *model.FundamentalsSnapshot / []model.NewsItem`。

### 新建:`model/scrape_convert.go`

8 個 converter,從 scrape DTO 直轉 model,語意對齊原 `emit.Map*DTO` 的輸出(key 命名 / Source 字串 / currency 規則 / 字串解析),讓 `tests/data_correctness_test.go` 端對端不變。

**共用 helper**(放同檔):
- `lineFromScaled(s *scrape.Scaled, currency string, ps, pe time.Time, key string) FundamentalsLine`——nil s 跳過;`Value: FromScaledDecimal(ScaledDecimal{Scaled: s.Scaled, Scale: int(s.Scale)})`;currency `strings.ToUpper` + 長度 3 檢查否則 `""`。
- `lineFromFloat(f *float64, currency string, ps, pe time.Time, key string, scale int) FundamentalsLine`——emit 把 `*float64` 乘 `10^scale` 進 ScaledDecimal;新 converter 直接 `Value: *f`(float64 語意等價,因 `FromScaledDecimal` 解碼後 = 原始 float64),省去縮放中介。
- `lineFromInt64(n *int64, ps, pe time.Time, key string) FundamentalsLine`——`Value: float64(*n)`,currency `""`。
- `pointInTimeBounds(asOf time.Time) (start, end time.Time)`——`start = AsOf 當天 00:00 UTC`,`end = start + 24h`(key-statistics / analysis / analyst-insights 共用)。
- `quarterBounds(asOf time.Time) (start, end time.Time)`——`qm = ((month-1)/3)*3+1`,`qs = Date(year, qm, 1, UTC)`,`end = qs.AddDate(0,3,-1)`(financials / balance-sheet / cash-flow 共用,對齊 emit `extractCurrentPeriodLines`)。
- `parseRevenueEstimateString(s string) (float64, bool)`——複製 emit 的邏輯(處理 "187.14B"/"1.2T"/"M"/"K" suffix、`$` 前綴、逗號),回傳 float64。
- `parseGrowthPercent(s string) (float64, bool)`——剝 `%` 後 `strconv.ParseFloat`。

**converter 簽名 + Source 字串 + key 對齊**(key 取自實讀 `emit.Map*DTO`):

| Converter | Source 字串 | Lines key(逐項對齊 emit) |
|---|---|---|
| `ScrapeFinancialsToSnapshot(dto *scrape.ComprehensiveFinancialsDTO, mic string) *FundamentalsSnapshot` | `yfinance/scrape/comprehensive-financials` | `extractCurrentPeriodLines` 的 22 key:`total_revenue`/`cost_of_revenue`/`gross_profit`/`operating_income`/`net_income`/`eps_basic`/`eps_diluted`/`ebit`/`ebitda`/`total_assets`/`total_debt`/`shareholders_equity`(=`common_stock_equity`)/`working_capital`/`tangible_book_value`/`operating_cash_flow`/`investing_cash_flow`/`financing_cash_flow`/`free_cash_flow`/`capital_expenditure`/`shares_outstanding_basic`(=`basic_average_shares`,int64)/`shares_outstanding_diluted`(=`diluted_average_shares`,int64)。時間窗用 `quarterBounds`。currency 用 `dto.Currency`。 |
| `ScrapeBalanceSheetToSnapshot(dto, mic)` | `yfinance/scrape/balance-sheet` | 5 key:`total_assets`/`total_debt`/`shareholders_equity`/`working_capital`/`tangible_book_value`。`quarterBounds`。`dto.Currency`。 |
| `ScrapeCashFlowToSnapshot(dto, mic)` | `yfinance/scrape/cash-flow` | 5 key:`operating_cash_flow`/`investing_cash_flow`/`financing_cash_flow`/`free_cash_flow`/`capital_expenditure`。`quarterBounds`。`dto.Currency`。 |
| `ScrapeKeyStatisticsToSnapshot(dto *scrape.ComprehensiveKeyStatisticsDTO, mic string)` | `yfinance/scrape/key-statistics` | 16 key:`market_cap`/`enterprise_value`(currency `dto.Currency`);`pe_ratio_trailing`/`pe_ratio_forward`/`peg_ratio`/`price_to_sales`/`price_to_book`/`ev_to_revenue`/`ev_to_ebitda`/`beta`/`profit_margin`/`operating_margin`/`return_on_assets`/`return_on_equity`(currency `""`);`shares_outstanding`(int64,currency `""`)。`pointInTimeBounds`。 |
| `ScrapeAnalysisToSnapshot(dto *scrape.ComprehensiveAnalysisDTO, mic string)` | `yfinance/scrape/analysis` | EPS 估計(`*float64`→`lineFromFloat`):`eps_estimate_current_quarter`/`_next_quarter`/`_current_year`/`_next_year`(currency `dto.EarningsEstimate.Currency`);`analyst_count_current_quarter`(int64,`""`);`eps_actual_recent`(float,currency `dto.EarningsHistory.Currency`);EPS trend:`eps_trend_current_quarter`/`_current_quarter_7d_ago`/`_current_quarter_30d_ago`/`_current_year`/`_next_year`(currency `dto.EPSTrend.Currency`);revisions:`eps_revisions_up_7d_current_quarter`/`_down_7d_`/`_up_30d_`/`_down_30d_`(int64,`""`);revenue estimate(字串解析):`revenue_estimate_current_quarter`/`_current_year`(currency `dto.RevenueEstimate.Currency`);`growth_estimate_current_year`(百分比字串解析,currency `""`)。`pointInTimeBounds`。 |
| `ScrapeAnalystInsightsToSnapshot(dto *scrape.AnalystInsightsDTO, mic string)` | `yfinance/scrape/analyst-insights` | price targets(`*float64`,currency `"USD"`):`current_price`/`target_price_mean`/`target_price_median`/`target_price_high`/`target_price_low`;`analyst_count`(int64,`""`);`recommendation_score`(`*float64`,currency `""`);`upside_potential_percent`(`(target-current)/current*100`,currency `""`)。`pointInTimeBounds`。 |
| `ScrapeNewsToItems(items []scrape.NewsItem) []NewsItem` | (N/A) | 每則:`Title`(trim)、`URL`(原樣,emit 的 `normalizeNewsURL`/`removeTrackingParams` 不重做——model 層不依賴 `net/url` 追蹤參數清理,留原 URL)、`Source`(`strings.TrimSpace`,emit 的 `cleanNewsSource` 字典不重做)、`Summary`(`""`,emit 也是空)、`PublishedAt`(nil → zero time,非 nil → `*PublishedAt.UTC()`;emit 的未來時間 clamp 不重做)、`Symbols`(`RelatedTickers` 原樣,emit 的去重+uppercase+長度驗證不重做——保持最小轉換)。nil/空 title/空 URL 的 item 跳過。 |

**所有 snapshot converter 共用**:`Symbol = dto.Symbol`、`MIC = mic`(caller 已 `inferMICForSymbol` 拿到,emit 的 `normalizeMIC` 不需複製)、`AsOf = dto.AsOf`。`ScrapeAllFundamentals` 不需新 converter,facade.Client fan-out 6 個 `Scrape*` 方法即可(語意等價)。

> 註:emit 的 `normalizeFinancialKey`(30 key 正規化,如 `total_revenues`→`total_revenue`)在 scrape DTO 來源已是 Yahoo canonical 命名,**新 converter 不再做此 mapping**,直接用上表 key。若 `tests/data_correctness_test.go` 對 key 有精確斷言,需確認(該測試目前只斷言 `len(Lines)>0` 與基本欄位,不檢查具體 key)。

### 刪除
- `svc/emit/` 全部 15 檔(10 非測試 + 5 測試):`bars.go`/`fundamentals.go`/`quotes.go`/`decimals.go`/`map_financials.go`/`map_news.go`/`map_profile.go`/`scrape_mapper.go`/`golden_marshaler.go`/`validation.go`/`canonical.go` + `decimals_test.go`/`golden_test.go`/`roundtrip_test.go`/`scrape_mapper_test.go`/`validation_test.go`。
- `tests/mapping/golden_test.go`(測 emit→proto round-trip)。
- `tests/crosslang/roundtrip_test.go`(測 emit→proto→python cross-lang)。

### 改寫
- `facade/client.go`:8 個 `Scrape*` 方法改走 `model.<converter>(dto, mic)`;刪 `svc/emit` import(L20),保留 `svc/scrape` import。範例:
  ```go
  // 改造前(ScrapeKeyStatistics):
  dto, err := scrape.ParseComprehensiveKeyStatistics(body, symbol, mic)
  snap, err := emit.MapKeyStatisticsDTO(dto, runID, "yfinance-go")
  return fromProtoFundamentals(snap), nil
  // 改造後:
  dto, err := scrape.ParseComprehensiveKeyStatistics(body, symbol, mic)
  return model.ScrapeKeyStatisticsToSnapshot(dto, mic), nil
  ```
  `ScrapeNews` 改:`return model.ScrapeNewsToItems(articles), nil`(注意 `scrape.ParseNews` 回傳 `([]scrape.NewsItem, *model.NewsStats, error)`,取第一個回傳值)。`runID` 參數在 converter 不再需要(proto Meta 欄位消失)——若要保留 runID lineage,可在 `model.FundamentalsSnapshot` 加 `RunID` 欄位,但當前 model 型別無此欄位,先不加(用戶本地匯出不需要 lineage)。
- `facade/fundamentals.go`:刪 `fromProtoFundamentals`(L65-106)+ `fundamentalsv1` import(L9);保留 `FundamentalsLine`/`FundamentalsSnapshot` alias 與 `FromFundamentalsSnapshot`(norm→SDK,不碰 proto)。
- `facade/news.go`:刪 `fromProtoNews`(L24-46)+ `newsv1` import(L8);保留 `NewsItem` alias。
- `cmd/scrape/scrape_run.go`:刪 `--preview-proto` 模式整段——`runScrapePreviewProto` 函式(L410-526)+ `PreviewProto` 欄位(L32)+ flag 綁定與說明(L43/52/62)+ dispatch 分支(L135-137)+ 錯誤訊息(L139)+ `validateScrapeFlags` 互斥檢查(L146)+ `--endpoints` 必填檢查(L194-196)+ `mapperConfig`(L425)+ 8 個 `emit.Map*` 呼叫(L453/463/471/479/489/499/507/515)+ `svc/emit` import(L18)。`--preview-json` / `--preview-news` / `--check` 路徑不動。
- `cmd/scrape/format.go`:刪 `printFundamentalsSnapshot`(L239)/`printProfileResult`(L258)/`printNewsArticles`(L270)/`getCurrencyFromLines`(L322)/`getTimeBounds`(L332)+ `fundamentalsv1`/`newsv1`/`svc/emit` import(L20-23);保留 `printAnalysisSummary`/`printAnalystInsightsSummary` 等 DTO-direct printer。
- `go.mod`:移除 `github.com/AmpyFin/ampy-proto/v2 v2.1.1`;`go mod tidy`——`google.golang.org/protobuf` 在 emit/bus 全清後無其他直接使用者(proto.Marshal 只在已刪的 bus publisher),會被 tidy 移除或降為 indirect。

### 不動
- `cmd/tools/golden/manifest_check.go`:獨立 main(不 import ampy-proto/emit,只用 schema 字串 `"ampy.bars.v1.BarBatch"` 等),不影響 yfin CLI build。
- `facade/samples/*`:註解提及 ampy-proto 但實際 import facade.* struct,不影響。
- `model/` 既有型別:`ScaledDecimal`/`FromScaledDecimal`/`InferMIC`/`FundamentalsLine`/`FundamentalsSnapshot`/`NewsItem`/`NewsStats` 等可重用元件保留。

### 驗證
```bash
go build ./... && go test ./...
go test ./tests/...                    # data_correctness_test.go 端對端
./yfin scrape --check AAPL --config config/effective.yaml
./yfin scrape --preview-json AAPL --endpoints key-statistics --config config/effective.yaml
./yfin scrape --preview-news AAPL --config config/effective.yaml
./yfin --help scrape                   # 不再列 --preview-proto
```

---

## 收尾驗證

```bash
go mod tidy
go build ./... && go test -race ./... && go vet ./...
grep -rn "ampy-proto\|ampy-bus\|ampy-config\|ampy-observability\|ampybus\|ampyobs\|ampyconfig\|fundamentalsv1\|newsv1\|commonv1\|svc/emit" --include='*.go' . | grep -v "cmd/tools/golden/manifest_check.go"
# 預期:0 hits(允許 manifest_check.go 的 schema 字串常數與註解殘留)
cat go.mod | grep -i ampy              # 預期:空
# 煙霧測試
./yfin pull AAPL --start 2024-01-01 --end 2024-01-31 --preview --config config/effective.yaml
./yfin quote AAPL --preview --config config/effective.yaml
./yfin scrape --preview-json AAPL --endpoints key-statistics,financials,news --config config/effective.yaml
./yfin admin config-effective --config config/effective.yaml
curl -s :9090/metrics | head            # Prometheus 仍正常
```

## 同步更新文件
- `CLAUDE.md`:依賴表移除四個 ampy* 行;「Why yahoo raw shapes...」段落若有提及 ampy-proto/emit 酌改。
- `README.md`:移除 `--publish` 提及(L622, L654)、ampy-proto/bus 整合段落、Scrape Fallback 中提及 ampy-proto 處。
- `docs/integrations/ampy-proto.md`:整檔刪或標示 deprecated。
- `docs/cli/usage.md` / `docs/cli/commands.md` / `docs/operations/error-handling.md`:移除 exit code 4 行、`--publish`/`--preview-proto` 說明。
- `docs/scrape/overview.md`:移除 emit/proto pipeline 描述,改為 scrape DTO → model 直通。

## 風險摘要
| 風險 | 緩解 |
|---|---|
| converter 與 emit 輸出語意不一致致 `data_correctness_test.go` 失敗 | Phase 4 完成後**先跑 `go test ./tests/...` 再跑其他**;key/Source/currency 嚴格對齊上表 |
| obsv OTLP exporter 構造失敗擋啟動 | fallback noop,不擋 |
| `google.golang.org/protobuf` 殘留為無用 direct | `go mod tidy` 處理 |
| cmd/samples 引用已刪 facade 函式 | facade.Client 公開 API(`Scrape*` 簽名)未變,不影響 |
| 4 phase 之間有隱藏耦合 | 每 phase 結束 `go build ./... && go test ./...` 全綠才進下一步 |
