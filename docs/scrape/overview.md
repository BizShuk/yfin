# 爬蟲備援系統概觀 (Scrape Fallback System Overview)

## 架構與資料流 (Architecture & Data Flow)

`yfin` 爬蟲備援系統提供一個 production-ready 的替代路徑，於 Yahoo Finance API 不可用、被限速、或需要付費訂閱時接手。整體設計以單一 `http.Client` + 完整重試/熔斷/限速策略為核心。

```tree
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   User Request  │    │  Orchestrator   │    │  Data Pipeline  │
│                 │    │                 │    │                 │
│ • CLI Command   │───▶│ • Rate Limiting │───▶│ • Normalization │
│ • Library Call  │    │ • Retry/Backoff │    │ • Validation    │
│ • Batch Job     │    │ • Circuit Break │    │ • Mapping       │
└─────────────────┘    └─────────────────┘    │ • Publishing    │
                                │              └─────────────────┘
                                ▼
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Data Sources  │    │     Output      │
                       │                 │    │                 │
                       │ • Yahoo API     │    │ • ampy-proto    │
                       │ • Web Scraping  │    │ • JSON Export   │
                       │ • FX Providers  │    │ • Bus Messages  │
                       └─────────────────┘    └─────────────────┘
```

## 為何需要爬蟲備援 (Why Scrape Fallback Was Created)

### 問題陳述 (Problem Statement)

Yahoo Finance API 有數項限制會直接衝擊 production 系統：

1. **認證要求**：多數進階 endpoint 需付費訂閱
2. **速率限制**：積極的節流策略會擋掉合法請求
3. **資料可用性**：部分資料只出現在網頁介面
4. **可靠性**：API endpoint 不穩定或暫時性失效

### 解法帶來的好處 (Solution Benefits)

- **自動備援**：API 與 scrape 之間無縫切換
- **資料一致性**：不同來源輸出一致
- **生產安全**：遵守 robots.txt，實施嚴格速率限制
- **完整覆蓋**：取得 API 無法取得的欄位

## 支援的資料來源 (Supported Data Sources)

### API Endpoints (主要)
- **Quotes**：即時與延遲報價
- **Historical Bars**：日/週/月 OHLCV
- **Fundamentals**：季度財報（需訂閱）

### Scrape Endpoints (備援)
- **Key Statistics**：P/E ratios、market cap、財務指標
- **Financials**：損益表、資產負債表、現金流量
- **Analysis**：分析師推薦與目標價
- **Profile**：公司資訊、經營階層、業務摘要
- **News**：近期新聞與新聞稿

## 資料流架構 (Data Flow Architecture)

### 1. 請求編排 (Request Orchestration)

```go
// 簡化版 orchestration 流程
func (o *Orchestrator) ProcessRequest(ctx context.Context, symbol string, endpoint string) (*Data, error) {
    strategy := o.determineStrategy(endpoint)

    switch strategy {
    case "api-first":
        data, err := o.tryAPI(ctx, symbol, endpoint)
        if shouldFallback(err) {
            return o.tryScrape(ctx, symbol, endpoint)
        }
        return data, err
    case "scrape-only":
        return o.tryScrape(ctx, symbol, endpoint)
    }
}
```

### 2. 資料正規化管線 (Data Normalization Pipeline)

```
Raw Data → Parser → Validator → Mapper → ampy-proto Message
    │         │         │         │            │
    ▼         ▼         ▼         ▼            ▼
  HTML/JSON  Regex   Schema    ScaledDec    UTC times
  嵌入 JSON  Extract Check     Currency     ISO codes
```

對應實作位置：

- `svc/scrape/` — HTML 解析、regex extractor、parser
- `svc/norm/` — 型別轉換、ScaledDecimal 包裝
- `svc/emit/` — DTO → ampy-proto mapper
- `utils/httpx` — 共用 HTTP client（限速/重試/熔斷）
- `utils/obsv` — observability bootstrap（metrics/tracing/logs）

## 資料契約與保證 (Data Contracts & Guarantees)

### 1. 時間標準化 (Time Standardization)
- 所有時間戳記皆為 UTC
- ISO-8601 格式
- 事件時間語意：bar close、新聞發布時間

### 2. 精度與貨幣 (Precision & Currency)
- ScaledDecimal：精確財務運算
- ISO-4217 貨幣代碼
- 價格、量、金額皆經 scaling 處理

### 3. 血緣與 metadata (Lineage & Metadata)
- Run ID tracking：每個請求都有唯一識別
- 來源標註：清楚的資料來源
- Schema 版本化：相容性管理

### 4. 錯誤處理 (Error Handling)
- 型別化錯誤：不同失敗模式對應特定 error type
- 優雅降級：部分資料仍勝過完全失敗
- 重試語意：智慧型 backoff retry

## 安全與合規 (Safety & Compliance)

### Robots.txt Compliance

系統支援三種執行層級（透過 `scrape.robots_policy`）：

1. **enforce**（生產環境）
   - 嚴格遵守 robots.txt 規則
   - 阻擋 disallowed 請求
   - 記錄違規事件

2. **warn**（開發環境）
   - 記錄 robots.txt 違規但放行
   - 用於測試階段

3. **ignore**（僅測試）
   - 完全跳過 robots.txt 檢查
   - 僅在受控測試環境使用

### Rate Limiting

- **Per-host QPS limit**：可設定請求速率（`scrape.qps`、`scrape.burst`）
- **Burst allowance**：處理瞬間流量
- **Exponential backoff**：智慧型重試時機（透過 `scrape.retry.{attempts,base_ms,max_delay_ms}`）

### Error Recovery

- **Circuit breaker**（`utils/httpx` 提供）：避免 cascade failure
- **Graceful degradation**：部分資料勝過完全失敗
- **Health check**：監控 endpoint 可用性
- **Automatic recovery**：狀態穩定後自動恢復正常操作

## 效能特性 (Performance Characteristics)

### Throughput Benchmarks

| Endpoint | API (req/s) | Scrape (req/s) | Fallback 額外成本 |
| --- | --- | --- | --- |
| Quotes | 10-15 | 2-3 | ~200ms |
| Key Stats | N/A | 1-2 | N/A |
| Financials | 5-8 | 1-2 | ~500ms |
| News | N/A | 0.5-1 | N/A |

### Latency Profiles

- API 請求：典型 100-500ms
- Scrape 請求：典型 800-2000ms
- Fallback 決策：<10ms
- 資料正規化：10-50ms

## 整合介面 (Integration Points)

### Library 用法

```go
import "github.com/bizshuk/yfin/svc/scrape"

client := scrape.NewClient(&scrape.Config{
    Enabled:   true,
    UserAgent: "AmpyFin-yfin/1.x",
    TimeoutMs: 10000,
    QPS:       0.7,
    Burst:     1,
    Retry:     scrape.RetryConfig{Attempts: 4, BaseMs: 300, MaxDelayMs: 4000},
    RobotsPolicy: "enforce",
    CacheTTLMs:   60000,
}, nil)

body, meta, err := client.Fetch(ctx, "https://finance.yahoo.com/quote/AAPL/key-statistics")
```

### CLI 用法

```bash
# 連線測試
yfin scrape --ticker AAPL --endpoint key-statistics --check

# extractor 乾跑
yfin scrape --ticker AAPL --endpoints key-statistics,financials,analysis,profile --preview-json

# proto 完整輸出乾跑
yfin scrape --ticker AAPL --endpoints financials,analysis,profile,news --preview-proto
```

## 監控與可觀測性 (Monitoring & Observability)

### 關鍵指標 (Key Metrics)

- **Fallback rate**：使用 scrape fallback 的請求比例
- **Success rate**：整體請求成功率
- **Latency percentiles**：P50、P95、P99
- **Error rates**：依 error type 與 endpoint 區分

### 健康指標 (Health Indicators)

- **Robots compliance**：enforce 模式下違規次數為零
- **Rate limit hits**：低於設定閾值
- **Parse success**：高抽取成功率
- **Schema validation**：所有輸出皆通過驗證

### 告警閾值 (Alerting Thresholds)

- 高 fallback rate：>50% 持續一段時間
- Parse failure：>5% 錯誤率
- Rate limit violation：生產環境中任何違規
- Schema drift：新的 parsing 錯誤

本概觀為理解爬蟲備援系統的基礎。詳細配置、使用範例、疑難排解請參考其他章節。