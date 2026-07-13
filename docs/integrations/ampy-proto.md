# AMPY-PROTO 整合指南

> 對應原文件：`docs/ampy-proto_scrapping.md` + `docs/scrapping.md` 的 AMPY-PROTO Integration 章節
>
> 本指南說明 `yfin`（SDK-first 重構後）如何將 Yahoo Finance 資料轉換為 AmpyFin 標準化 `ampy-proto` 訊息，並透過 `utils/bus` 發佈至 AmpyFin 生態系的下游系統。

---

## Overview

`yfin` 將所有 8 個 scrape endpoint + chart API endpoint 都轉成 `ampy-proto/v2` 訊息，提供 AmpyFin 生態系的標準化資料介面。

### 什麼是 AMPY-PROTO

`ampy-proto` 是 AmpyFin 制定的金融資料通訊協定，提供：

- 標準化訊息格式 (`Standardized Message Formats`)：所有金融資料型別共用一致的結構
- 高精度小數 (`Scaled Decimal Precision`)：以 `ScaledDecimal` 避免 IEEE-754 浮點誤差
- 豐富的 metadata：RunID、timestamp、schema version、source、producer 等觀測欄位
- 安全識別 (`Security Identification`)：統一的 `Symbol` + `Market Identifier Code (MIC)` 編碼
- 週期性資料 (`Period-based Data`)：以 `PeriodStart` / `PeriodEnd` 框定財務期間

### 誰會消費

- AmpyFin trading platform（訊號產生、部位管理、風控引擎）
- 下游資料倉儲（ClickHouse / BigQuery 等的 `ampy-proto` schema loader）
- 報表服務（risk dashboard、P&L attribution、news sentiment）

---

## Pipeline

`yfin` 的 scrape → proto 流程分為四個階段：raw HTTP → `svc/yahoo` 或 `svc/scrape` → `svc/norm` → `svc/emit` → `utils/bus` publish。

```text
┌─────────────────┐
│  Yahoo Finance  │
│  (chart / HTML) │
└────────┬────────┘
         │
         ▼
┌─────────────────┐  ┌─────────────────┐
│  svc/yahoo      │  │  svc/scrape     │
│  (chart API)    │  │  (HTML scrape)  │
└────────┬────────┘  └────────┬────────┘
         │                    │
         ▼                    ▼
┌─────────────────────────────────────────┐
│  svc/norm                               │
│  (YahooDTO / ScrapeDTO → NormalizedDTO) │
└────────────────────┬────────────────────┘
                     │ ScaledDecimal precision
                     ▼
┌─────────────────────────────────────────┐
│  svc/emit                               │
│  EmitBarBatch / EmitQuote /             │
│  EmitFundamentals / ScrapeMapper.*      │
└────────────────────┬────────────────────┘
                     │ ampy-proto/v2 proto
                     ▼
┌─────────────────────────────────────────┐
│  utils/bus                              │
│  BusPublisher.PublishBars / Quote /     │
│  Fundamentals → ampy-bus envelope       │
└────────────────────┬────────────────────┘
                     │ NATS subject / Kafka topic
                     ▼
              ampy.bus.stream
```

四階段職責：

| 階段 | 套件 | 職責 |
| --- | --- | --- |
| Fetch | `svc/yahoo` / `svc/scrape` | chart API 或 HTML scrape，回傳 raw DTO |
| Normalize | `svc/norm` | 統一欄位命名、補上 MIC、`ScaledDecimal` 化價格/金額 |
| Emit | `svc/emit` | 將 `norm.Normalized*` 轉成 `ampy-proto/v2` 訊息 |
| Publish | `utils/bus` | 套上 envelope + chunking + retry/CB 後丟到 NATS / Kafka |

---

## Message Schema 概覽

`yfin` 目前產生下列 `ampy-proto/v2` 訊息型別：

| 訊息型別 | 用途 | 來源 |
| --- | --- | --- |
| `ampy.bars.v1.BarBatch` | 每日 OHLC bars（含 adjustment policy） | `svc/emit.EmitBarBatch` |
| `ampy.ticks.v1.QuoteTick` | 即時報價（bid / ask / venue） | `svc/emit.EmitQuote` |
| `ampy.fundamentals.v1.FundamentalsSnapshot` | 財報三表 + 統計 + 分析 + 分析師洞察 | `svc/emit.EmitFundamentals`、`svc/emit.ScrapeMapper.MapFinancials` |
| `ampy.profile.v1.ProfileSnapshot` | 公司基本資料（fallback 為 canonical JSON） | `svc/emit.ScrapeMapper.MapProfile`、`svc/emit.MapProfileDTO` |
| `ampy.news.v1.NewsItem` | 新聞文章 metadata | `svc/emit.ScrapeMapper.MapNews`、`svc/emit.MapNewsItems` |

所有訊息共用 `common.v1.SecurityId { Symbol, Mic }` 與 `common.v1.Meta { RunId, Source, Producer, SchemaVersion, ProducedAt }`。

### ScaledDecimal 編碼

價格與金額以 `ScaledDecimal` 內部儲存；`ampy-proto` 以 `common.v1.Decimal { Scaled int64, Scale int32 }` 表示：

```text
394,328,000,000    (scale 0) →  $394.328 B  revenue
       3,175       (scale 2) →  31.75        forward P/E
       2,430       (scale 2) →  24.30 %      profit margin
```

換算公式：

```go
func fromScaled(scaled int64, scale int32) float64 {
    return float64(scaled) / math.Pow10(int(scale))
}
```

scale 合法範圍為 `0..9`（於 `svc/emit.decimals.go` 的 `ValidateDecimal` 強制檢查）。

### 範例：FundamentalsSnapshot

```json
{
  "security": {"symbol": "AAPL", "mic": "XNAS"},
  "meta": {
    "runId": "run_2025_01_29_17_23_27",
    "source": "yfin/scrape",
    "producer": "yfin",
    "schemaVersion": "ampy.fundamentals.v1:2.1.0",
    "producedAt": "2025-01-29T17:23:27Z"
  },
  "lines": [
    {
      "key": "total_revenue",
      "value": {"scaled": 394328000000, "scale": 0},
      "currencyCode": "USD",
      "periodStart": "2024-10-01T00:00:00Z",
      "periodEnd":   "2024-12-31T23:59:59Z"
    }
  ],
  "source": "yfin/scrape/financials",
  "asOf": "2024-12-31T23:59:59Z"
}
```

### 範例：Bar

```json
{
  "security": {"symbol": "AAPL", "mic": "XNAS"},
  "start":  "2024-12-30T00:00:00Z",
  "end":    "2024-12-30T23:59:59Z",
  "open":   {"scaled": 25140, "scale": 2},
  "high":   {"scaled": 25230, "scale": 2},
  "low":    {"scaled": 25090, "scale": 2},
  "close":  {"scaled": 25198, "scale": 2},
  "volume": 52340000,
  "adjusted": true,
  "adjustmentPolicyId": "split_dividend",
  "adjustmentPolicy":  "ADJUSTMENT_POLICY_SPLIT_DIVIDEND"
}
```

---

## Producer / Source 慣例

`yfin` 在所有 emit/bus metadata 採用兩段式標識，方便下游篩選與歸因：

| 欄位 | 慣例值 | 說明 |
| --- | --- | --- |
| `producer` | `yfin` | 標識「哪個 SDK / 應用」產生；固定為 `yfin` |
| `source` | `yfin/<sub-system>` | 標識「SDK 內哪條路徑」產生，例如 `yfin/scrape`、`yfin/yahoo`、`yfin/twse` |

### 各 emit 路徑的 source 值

| Emit 入口 | `source` 值 |
| --- | --- |
| `svc/emit.EmitBarBatch` | `yfin/yahoo` |
| `svc/emit.EmitQuote` | `yfin/yahoo` |
| `svc/emit.EmitFundamentals`（chart API 來源） | `yfin/yahoo` |
| `svc/emit.ScrapeMapper.MapFinancials` | `yfin/scrape/financials` |
| `svc/emit.ScrapeMapper.MapProfile` | `yfin/scrape/profile` |
| `svc/emit.ScrapeMapper.MapNews` | `yfin/scrape/news` |
| `svc/emit.MapFinancialsDTO` / `MapComprehensiveFinancialsDTO` | `yfin/scrape/<endpoint>`（endpoint-specific） |
| `svc/emit.MapProfileDTO` | `yfin/scrape/profile` |
| `svc/emit.MapNewsItems` | `yfin/scrape/news` |

> SDK-first 重構期間，舊版程式碼仍以舊字串標識；本指南以重構後的目標值 `yfin` / `yfin/<sub>` 為準。

`producer` 與 `source` 差異：

- 多 `yfin` instance 同時運行時，`source` 可幫助定位「哪條子系統路徑」產出。
- `producer` 主要做 SDK-level 歸因、版本追蹤；envelope 上的 producer header 在 `utils/bus` 端會組合成 `yfin@<hostname>` 形式供下游 routing。

---

## Per-endpoint Mapping

### Scrape endpoints（HTML → proto）

`svc/scrape` 的 8 個 endpoint 全部對應 `ampy-proto` 訊息：

| Scrape endpoint | Proto message | Emit 入口 |
| --- | --- | --- |
| `profile` | `ampy.profile.v1.ProfileSnapshot` | `ScrapeMapper.MapProfile` / `MapProfileDTO` |
| `key-statistics` | `ampy.fundamentals.v1.FundamentalsSnapshot` | `MapFinancialsDTO` / `MapComprehensiveFinancialsDTO` |
| `financials` | `ampy.fundamentals.v1.FundamentalsSnapshot` | 同上 |
| `balance-sheet` | `ampy.fundamentals.v1.FundamentalsSnapshot` | 同上 |
| `cash-flow` | `ampy.fundamentals.v1.FundamentalsSnapshot` | 同上 |
| `analysis` | `ampy.fundamentals.v1.FundamentalsSnapshot` | 同上 |
| `analyst-insights` | `ampy.fundamentals.v1.FundamentalsSnapshot` | 同上 |
| `news` | `ampy.news.v1.NewsItem` | `ScrapeMapper.MapNews` / `MapNewsItems` |

CLI 用法（AMPY-PROTO preview 模式）：

```bash
yfin scrape --preview-proto --ticker AAPL \
  --endpoints financials,balance-sheet,cash-flow,key-statistics,analysis,analyst-insights,profile,news \
  --config config/effective.yaml
```

單一 endpoint：

```bash
yfin scrape --preview-proto --ticker AAPL --endpoints financials --config config/effective.yaml
yfin scrape --preview-proto --ticker AAPL --endpoints profile     --config config/effective.yaml
yfin scrape --preview-proto --ticker AAPL --endpoints news        --config config/effective.yaml
```

### Chart API endpoints（Yahoo Finance API → proto）

| Yahoo endpoint | Proto message | Emit 入口 |
| --- | --- | --- |
| `/v8/finance/chart/{symbol}`（daily） | `ampy.bars.v1.BarBatch` | `svc/emit.EmitBarBatch` |
| `/v6/finance/quote`（含 `quoteSummary`） | `ampy.ticks.v1.QuoteTick` | `svc/emit.EmitQuote` |
| `/v10/finance/quoteSummary` | `ampy.fundamentals.v1.FundamentalsSnapshot` | `svc/emit.EmitFundamentals` |

CLI 用法：

```bash
# 每日 bars（chart API → ampy.bars.v1.BarBatch → bus）
yfin pull   --ticker AAPL --start 2024-01-01 --end 2024-12-31 --publish --config config/effective.yaml

# 即時 quote（quoteSummary → ampy.ticks.v1.QuoteTick → bus）
yfin quote  --ticker AAPL --publish --config config/effective.yaml
```

### Line item 名稱對照（scrape → proto）

`Svc/emit.MapFinancialsDTO` 將 scrape 解析出的 `Key` 標準化為下列 snake_case line items：

| Endpoint | Line items |
| --- | --- |
| `financials` | `total_revenue`, `operating_income`, `net_income`, `ebitda`, `basic_eps`, `diluted_eps` |
| `balance-sheet` | `total_assets`, `total_debt`, `shareholders_equity`, `working_capital`, `tangible_book_value`, `cash_and_equivalents` |
| `cash-flow` | `operating_cash_flow`, `investing_cash_flow`, `financing_cash_flow`, `free_cash_flow`, `capital_expenditure` |
| `key-statistics` | `market_cap`, `enterprise_value`, `forward_pe`, `trailing_pe`, `peg_ratio`, `price_sales`, `price_book`, `beta`, `shares_outstanding`, `profit_margin`, `operating_margin`, `return_on_assets`, `return_on_equity` |
| `analysis` | `eps_estimate_current_year`, `eps_estimate_next_year`, `revenue_estimate_current_year`, `revenue_estimate_next_year`, `growth_estimate_current_year`, `growth_estimate_next_year` |
| `analyst-insights` | `price_target_high`, `price_target_low`, `price_target_median`, `price_target_average`, `recommendation_score`, `number_of_analysts` |

---

## Topic Naming

`utils/bus.TopicBuilder` 採用 `<prefix>.<env>.<domain>.<version>.<subtopic>` 五段式結構：

```text
prefix    env     domain           version  subtopic
──────    ───     ──────           ───────  ────────
ampy  .   dev  .  bars        .    v1   .  XNAS.AAPL
ampy  .   dev  .  ticks       .    v1   .  XNAS.AAPL
ampy  .   dev  .  fundamentals.    v1   .  AAPL       ← MIC 省略，只用 symbol
ampy  .   dev  .  news        .    v1   .  AAPL
```

### Topic 各段說明

| 段 | 來源 | 說明 |
| --- | --- | --- |
| `prefix` | `bus.topic_prefix`（YAML `bus.topic_prefix`，預設 `ampy`） | 整個 AmpyFin 生態系的 namespace prefix |
| `env` | `app.env`（YAML `app.env`，`dev` / `staging` / `prod`） | 環境隔離，避免 dev/staging/prod 互相干擾 |
| `domain` | 訊息種類決定 | 合法值：`bars`、`ticks`、`fundamentals`、`news`、`fx`、`signals`、`orders`、`fills`、`positions`、`metrics`、`dlq`、`control` |
| `version` | 固定 `v1`（當前） | schema 版本；後續 breaking change 升 `v2` |
| `subtopic` | `bus.Key` 決定 | bars / ticks 採 `<MIC>.<symbol>`；fundamentals / news 採 `<symbol>` |

`utils/bus.TopicBuilder.ValidateTopic` 強制 topic 至少 4 段且 `domain` 為合法值、`version` 必須 `v` 開頭。

### NATS subject 對應

`BusPublisher.createNATSBus` 將 `<prefix>.<env>.>` 訂閱為 single subject tree，便於 consumer 用單一訂閱接收整個環境下所有 `yfin` 訊息：

```text
NATS subjects:
  ampy.dev.bars.v1.XNAS.AAPL
  ampy.dev.ticks.v1.XNAS.AAPL
  ampy.dev.fundamentals.v1.AAPL
  ampy.dev.news.v1.AAPL

Consumer subscribes to:  ampy.dev.>
```

### Config 區塊

`config/effective.yaml` 的 `bus` 區塊控制所有上述行為：

```yaml
bus:
  enabled: false                  # 全域開關；關閉時 Publish* 直接 error
  env: "dev"                      # topic env segment
  topic_prefix: "ampy"            # topic prefix
  max_payload_bytes: 1048576      # 256 KiB ~ 10 MiB（utils/bus.ValidateConfig 強制範圍）
  publisher:
    backend: "nats"               # nats | kafka（kafka 尚未實作）
    nats:
      url: "${NATS_URL:-nats://localhost:4222}"
      subject_style: "topic"
      ack_wait_ms: 5000
    kafka:
      brokers: []
      acks: "all"
      compression: "snappy"
  retry:
    attempts: 5
    base_ms: 250
    max_delay_ms: 8000
  circuit_breaker:
    window: 50
    failure_threshold: 0.30
    reset_timeout_ms: 30000
    half_open_probes: 3
```

完整設定請見 `docs/operations/configuration.md`。

---

## Sample Pipeline Code

下例展示從 normalized DTO 一路 emit 到 `utils/bus` publish 的完整鏈路：

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/bizshuk/yfin/facade"
    "github.com/bizshuk/yfin/svc/emit"
    "github.com/bizshuk/yfin/svc/norm"
    "github.com/bizshuk/yfin/utils/bus"
)

func publishBars(ctx context.Context, symbol string) error {
    // 1. 用 facade 抓取 daily bars（回傳 norm.NormalizedBarBatch）
    client, err := facade.NewClient()
    if err != nil {
        return fmt.Errorf("create client: %w", err)
    }
    batch, err := client.FetchDailyBarsNorm(ctx, symbol, "2024-01-01", "2024-12-31", true, "")
    if err != nil {
        return fmt.Errorf("fetch bars: %w", err)
    }

    // 2. emit → ampy.bars.v1.BarBatch
    ampyBatch, err := emit.EmitBarBatch(batch)
    if err != nil {
        return fmt.Errorf("emit bars: %w", err)
    }

    // 3. 透過 utils/bus publish
    busCfg := bus.GetConfigFromEnv()
    if !busCfg.Enabled {
        log.Println("bus disabled; skip publish")
        return nil
    }
    busInst, err := bus.NewBus(busCfg)
    if err != nil {
        return fmt.Errorf("create bus: %w", err)
    }
    defer busInst.Close(ctx)

    msg := &bus.BarBatchMessage{
        Batch: ampyBatch,
        Key: &bus.Key{
            Symbol: batch.Security.Symbol,
            MIC:    batch.Security.MIC,
        },
        RunID: batch.Meta.RunID,
        Env:   busCfg.Env,
    }

    if err := busInst.PublishBars(ctx, msg); err != nil {
        return fmt.Errorf("publish bars: %w", err)
    }
    log.Printf("published %d bars for %s", len(batch.Bars), symbol)
    return nil
}
```

Scrape 路徑（`svc/scrape` → `svc/emit.ScrapeMapper` → bus）：

```go
// 從 svc/scrape 拿到 financials DTO
dto, fetchMeta, err := client.ScrapeFetch(ctx, "AAPL", scrape.EndpointFinancials)
if err != nil { return err }

// 透過 ScrapeMapper 轉 proto
mapper := emit.NewScrapeMapper(emit.ScrapeMapperConfig{
    RunID:    runID,
    Producer: "yfin",
    Source:   "yfin/scrape",
    TraceID:  traceID,
})
snapshot, err := mapper.MapFinancials(ctx, dto)
if err != nil { return err }

// bus publish
msg := &bus.FundamentalsMessage{
    Fundamentals: snapshot,
    Key:          &bus.Key{Symbol: dto.Symbol, MIC: dto.Market},
    RunID:        runID,
    Env:          busCfg.Env,
}
return busInst.PublishFundamentals(ctx, msg)
```

實際 CLI 入口請見 `cmd/market/pull.go`（`PublishBars`）與 `cmd/market/quote.go`（`PublishQuote`）。

---

## See also

- `docs/operations/configuration.md` — `bus:` 區塊完整設定（backend、retry、CB、NATS/Kafka）
- `docs/api/data-structures.md` — `facade.Bar` / `facade.Quote` / `facade.FundamentalsSnapshot` 等對外公開 struct
- `docs/cli/usage.md` — `--publish` / `--preview` 旗標說明（bars、quote、scrape）
- `docs/scrape/overview.md` — scrape 8 個 endpoint 的 parser 細節與 robots.txt 行為
- `svc/emit/` — emit 套件（`EmitBarBatch`、`EmitQuote`、`EmitFundamentals`、`ScrapeMapper`）
- `svc/norm/` — `NormalizedBarBatch` / `NormalizedQuote` / `NormalizedFundamentalsSnapshot` 型別定義
- `utils/bus/` — topic builder、envelope、retry/CB、publisher