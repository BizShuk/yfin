# yfinance-go 架構與套件指南 (Architecture and Packages Guide)

本教學文件將逐一介紹 `yfinance-go` 專案中的所有套件 (package)。本專案採用模組化設計，將職責清晰劃分，確保系統的強健性 (robustness)、可觀測性 (observability) 與高精度的資料正規化。

---

## 🗺️ 系統架構關聯圖

以下為專案中各套件的依賴與呼叫關係圖：

```mermaid
flowchart TD
    CMD["cmd/yfin (命令列進入點)"] -->|"呼叫"| YF["yfinance (高階用戶端)"]
    CMD -->|"讀取"| CFG["internal/config (組態管理)"]
    CMD -->|"查詢頻率"| CCH["internal/cache (快取機制)"]

    EXT["外部消費者 (External Consumers)"] -->|"整合"| FAC["facade (外觀層)"]
    FAC -->|"轉換資料"| YF

    YF -->|"原始 API 請求"| YAH["internal/yahoo (原始 API 用戶端)"]
    YF -->|"爬網備援"| SCR["internal/scrape (網頁爬蟲引擎)"]
    YF -->|"正規化轉換"| NORM["internal/norm (資料正規化)"]
    YF -->|"台灣市場"| TWSE["svc/twse (台灣證券交易所服務)"]

    SOAK -->|"使用"| YF

    YAH -->|"基礎 HTTP 傳輸"| HTX["internal/httpx (彈性 HTTP 用戶端)"]
    SCR -->|"基礎 HTTP 傳輸"| HTX

    HTX -->|"記錄指標"| OBS["internal/obsv (可觀測性)"]

    NORM -->|"映射與校驗"| EMT["internal/emit (資料映射與驗證)"]
    EMT -->|"發佈訊息"| BUS["internal/bus (匯流排發佈)"]

    SOAK["cmd/soak (壓力測試獨立 binary)"] -->|"使用"| YF
```

---

## 1. `yfinance` (根目錄套件)

### 📌 職責說明

`yfinance` 套件位於專案根目錄，是本專案的進入點 (entry point) 與對外的高階用戶端 (high-level client)。

### 📁 關鍵檔案

- `client.go`: 定義 `Client` 結構，實現核心 fetch 方法，並處理備援 (fallback) 機制。

### ⚙️ 套件設計與運作

- 整合了 `internal/yahoo` 的原始 API 呼叫與 `internal/scrape` 的爬蟲備援引擎。
- 當付費 API 限制、網路逾時 (timeouts) 或回傳 `429` 錯誤時，能自動且無縫地將請求遞補至爬網流程。
- 所有取回的資料會經由 `internal/norm` 正規化後，以統一的格式回傳給呼叫端。

---

## 2. `facade`

### 📌 職責說明

`facade` 套件實作了外觀模式 (facade pattern)，對外暴露反射免除 (reflection-free) 的純 Go 結構 (plain Go structs)。

### 📁 關鍵檔案

- `bars.go`: 轉換並包裝 K 線 (bar) 資料。
- `quote.go`: 轉換並包裝即時報價 (quote) 資料。
- `company_info.go`: 轉換並包裝公司基本資訊 (company info)。

### ⚙️ 套件設計與運作

- 外部套件（例如 `stock` 或是 `data`）在匯入本專案時，應以此套件作為邊界，避免直接存取 `internal/` 的複雜結構。
- 將內部的定點十進位 `ScaledDecimal` 轉化為標準的 `float64` 浮點數，便於資料序列化 (serialization) 與一般的數值計算。

---

## 3. `cmd/yfin`

### 📌 職責說明

`cmd/yfin` 是命令列介面的主要入口 (CLI entry point) 與組合根 (composition root)。

### 📁 關鍵檔案

- `main.go`: 系統主程式入口。
- `dispatch.go`: 註冊並分派所有抓取指令 (command registry)。
- `batch.go`, `pull.go`, `soak.go`: 實作特定的 CLI 命令子集。

### ⚙️ 套件設計與運作

- 使用 `cobra` 套件管理命令列參數。
- 解析使用者指定的 YAML 設定檔，初始化連線池，並根據輸入的指令執行批次下載 (`batch`)、單股抓取 (`pull`) 或進行壓力測試 (`soak`)。

---

## 4. `internal/bus`

### 📌 職責說明

`internal/bus` 負責將下載完成的正規化金融資料發佈 (publish) 至 `ampy-bus` 訊息系統。

### 📁 關鍵檔案

- `bus.go`: 定義匯流排的介面與基本功能。
- `publisher.go`: 實作具備自動重試功能的發佈器 (publisher)。
- `chunking.go`: 當資料批次過大時，自動分切為較小的區塊 (chunks)。
- `envelope.go`: 包裝資料信封 (envelope) 並注入追蹤用元資料 (metadata)。

### ⚙️ 套件設計與運作

- 提供底層發佈通道，支援異步發佈 (asynchronous publishing)。
- 在網路不穩定時，內建重試與退避邏輯，確保資料不遺失地傳遞到其他微服務。

---

## 5. `internal/cache`

### 📌 職責說明

`internal/cache` 實作了專屬於 Yahoo Finance 資料抓取的更新頻率快取機制 (caching mechanism)。

### 📁 關鍵檔案

- `refresh.go`: 定義不同資料維度的快取保留期 (retention period)。
- `tickerlist.go`: 處理本地端 Ticker 清單的讀取與載入。

### ⚙️ 套件設計與運作

- 依據資料屬性定義更新間隔：如 `daily` (每日)、`monthly` (每月) 與 `quarterly` (每季)。
- 提供 `REFRESH_MAP` 規則引擎，避免對不常變動的資料（如公司 Profile、Holder 結構）頻繁進行網路請求。

---

## 6. `internal/config`

### 📌 職責說明

`internal/config` 負責解析並驗證系統的組態設定 (configuration)。

### 📁 關鍵檔案

- `ampy_config.go`: 解析並驗證 `ampy-config` YAML 的系統設定。
- `httpclient.go`: 提供 HTTP 用戶端的組態解析結構。

### ⚙️ 套件設計與運作

- 在系統啟動時載入組態檔案，檢查並行工作線數 (concurrency workers)、QPS 限流速率與斷路器參數之合理性。
- 將設定結果傳遞給 `internal/httpx` 作為初始化基礎。

---

## 7. `internal/emit`

### 📌 職責說明

`internal/emit` 負責將正規化的資料格式化為標準的 `ampy-proto` 定義，並在發出前執行強健性驗證 (robust validation)。

### 📁 關鍵檔案

- `decimals.go`: 提供定點十進位 `ScaledDecimal` 的轉換與溢位檢查。
- `validation.go`: 進行資料邏輯檢驗（例如確保低價 `low` 不大於收盤價 `close` 與開盤價 `open`）。
- `map_financials.go`, `map_news.go`, `map_profile.go`: 對 scraped 欄位進行 canonical 欄位對齊映射。

### ⚙️ 套件設計與運作

- 充當系統向外部輸出 Protobuf 資料的前哨站，保障流入訊息匯流排的資料符合協議版本與邊界約束。
- 當發現 OHLC 價格存在細微的浮點數誤差或異常時，此套件會進行自動微調（例如當 `low` 因浮點數極小誤差大於 `close` 時，自動調整其值），以避免下游解碼失敗。

---

## 8. `internal/httpx`

### 📌 職責說明

`internal/httpx` 提供一個具備高彈性 (resilience) 的 HTTP 用戶端。

### 📁 關鍵檔案

- `client.go`: 核心用戶端，整合限流、重試與斷路器。
- `limiter.go`: 實作權杖桶 (token bucket) 演算法的速率限制器 (rate limiter)。
- `circuit_breaker.go`: 實作斷路器 (circuit breaker)，在遠端伺服器大量發生故障時暫時熔斷請求。
- `errors.go`: 定義連線、超時與限流之自訂錯誤型態。

### ⚙️ 套件設計與運作

- 採用單一共享的 `http.Client` 以最大化重複利用連線（基於 `Keep-Alive` 協議）。
- 整合了 QPS 速率限制，當 API 回傳 `429` 或是偵測到失敗率過高時，利用指數退避 (exponential backoff) 與隨機抖動 (jitter) 進行延遲重試。
- 當連續請求失敗次數到達閾值，斷路器進入 `open` (開啟) 狀態，立即拒絕後續呼叫以保護用戶端，直到冷卻重置時間結束。

---

## 9. `internal/norm`

### 📌 職責說明

`internal/norm` 負責將抓取到的原始半結構化或非結構化資料，轉換並歸一化為系統標準領域模型 (domain model)。

### 📁 關鍵檔案

- `conversion.go`: 原始數值與型態之強型態安全轉換器。
- `security.go`: 推斷與補齊 Market Identifier Code (MIC) 及交易所屬性。
- `decimal.go`: 處理 `ScaledDecimal` 十進位轉換細節。
- `time.go`: 將多元時區之時間戳記統一轉換為 `UTC` ISO-8601 格式。

### ⚙️ 套件設計與運作

- 處理來自 Yahoo JSON API 的複雜欄位，如 `regularMarketPrice` 與各種時間欄位。
- 在 MIC 無法由 API 取得時，根據 symbol 尾碼（如 `.TW` 代表台灣證券交易所，MIC 為 `XTAI`）進行靜態推導與動態快取，確保資料識別資訊完整。

---

## 10. `internal/obsv`

### 📌 職責說明

`internal/obsv` 負責整合系統的可觀測性 (observability) 架構，包含度量指標 (metrics) 與追蹤 (tracing)。

### 📁 關鍵檔案

- `obsv.go`: 提供 OpenTelemetry 與追蹤監控的初始化方法。
- `metrics.go`: 包裝 Prometheus 註冊器與度量計數指標。

### ⚙️ 套件設計與運作

- 對 HTTP 用戶端的每次傳輸過程進行攔截，收集延遲時間、成功/失敗率、速率限制命中數。
- 將可觀測性資料匯出至 `inf` 監控後端（例如 Prometheus 與 VictoriaMetrics）。

---

## 11. `internal/scrape`

### 📌 職責說明

`internal/scrape` 是網頁爬網引擎 (scraping engine)，提供在 API 故障或需存取付費欄位時的替代解決方案。

### 📁 關鍵檔案

- `client.go`: 網頁抓取的核心控制流程。
- `robots.go`: 用於解析與嚴格遵守遠端網站 `robots.txt` 爬蟲協定的模組。
- `financials.go`, `analysis.go`, `statistics.go`: 負責解析特定 HTML/JSON 網頁節點。
- `extract_news.go`: 新增的 JSON 新聞擷取模組。

### ⚙️ 套件設計與運作

- 在發起抓取前，會先行請求並快取 `robots.txt`，確保所有請求路徑均符合合規性標準。
- 專門擷取分析師預測 (`ScrapeAnalysis`)、詳細財務報表 (`ScrapeFinancials`) 與即時新聞內容 (`ScrapeNews`)，彌補免費 API 的欄位缺憾。

---

## 12. `cmd/soak`

### 📌 職責說明

`cmd/soak` 負責長時間穩定性測試 (soak testing) 的調度與執行，是獨立的 CLI binary。

### 📁 關鍵檔案

- `orchestrator.go`: 壓測工作的協調器。
- `worker.go`: 並行拉取工作單元。
- `probes.go`: 數值與邏輯正確性探針。
- `memory.go`: 定期調用 runtime 進行記憶體使用增長與洩漏分析。

### ⚙️ 套件設計與運作

- 在開發與持續整合 (CI) 階段，被用來進行高強度壓測，長時間以並行 goroutine 拉取真實市場資料。
- 正確性探針會對產出的 `NormalizedBar` 進行極限值、時間順序性等比對，並產出壓測報告。

---

## 13. `internal/yahoo`

### 📌 職責說明

`internal/yahoo` 負責與 Yahoo Finance 的原始 HTTP API 介面進行通訊。

### 📁 關鍵檔案

- `client.go`: 對接原始 API (如 `/v10/finance/quoteSummary`, `/v7/finance/options`)。
- `auth.go`: 管理 Yahoo API 認證所需的 `Cookie` 與 `Crumb` 機制。
- `bars.go`, `quotes.go`, `fundamentals.go`: 對接並解析 API 報文 (raw responses)。

### ⚙️ 套件設計與運作

- 由於 Yahoo Finance 頻繁變更 Crumb 授權機制，此套件負責實作 Crumb 的抓取與快取。
- 利用 `internal/httpx` 提供連線，以取得最原始的、未加工過的 Yahoo 資料結構。

---

## 14. `svc/twse`

### 📌 職責說明

`svc/twse` 為台灣市場專屬的證券交易所 (Taiwan Stock Exchange) 數據處理服務。

### 📁 關鍵檔案

- `client.go`: TWSE 資料獲取用戶端。
- `parser.go`: 解析 TWSE 官方公佈的 BWIBBU (個股本益比、殖利率及股淨比) 與 T86 (三大法人買賣超) 開放資料。

### ⚙️ 套件設計與運作

並行拉取工作單元。

- `probes.go`: 數值與邏輯正確性探針。
- `memory.go`: 定期調用 runtime 進行記憶體使用增長與洩漏分析。

### ⚙️ 套件設計與運作

- 在開發與持續整合 (CI) 階段，被用來進行高強度壓測，長時間以並行 goroutine 拉取真實市場資料。
- 正確性探針會對產出的 `NormalizedBar` 進行極限值、時間順序性等比對，並產出壓測報告。

---

## 13. `internal/yahoo`

### 📌 職責說明

`internal/yahoo` 負責與 Yahoo Finance 的原始 HTTP API 介面進行通訊。

### 📁 關鍵檔案

- `client.go`: 對接原始 API (如 `/v10/finance/quoteSummary`, `/v7/finance/options`)。
- `auth.go`: 管理 Yahoo API 認證所需的 `Cookie` 與 `Crumb` 機制。
- `bars.go`, `quotes.go`, `fundamentals.go`: 對接並解析 API 報文 (raw responses)。

### ⚙️ 套件設計與運作

- 由於 Yahoo Finance 頻繁變更 Crumb 授權機制，此套件負責實作 Crumb 的抓取與快取。
- 利用 `internal/httpx` 提供連線，以取得最原始的、未加工過的 Yahoo 資料結構。

---

## 14. `svc/twse`

### 📌 職責說明

`svc/twse` 為台灣市場專屬的證券交易所 (Taiwan Stock Exchange) 數據處理服務。

### 📁 關鍵檔案

- `client.go`: TWSE 資料獲取用戶端。
- `parser.go`: 解析 TWSE 官方公佈的 BWIBBU (個股本益比、殖利率及股淨比) 與 T86 (三大法人買賣超) 開放資料。

### ⚙️ 套件設計與運作

- 由於部分台灣市場指標（如法人籌碼面、本益比歷史明細）在 Yahoo Finance 較難完整取得，此套件直接請求 TWSE Open Data。
- 經過專屬解析器後，將 TWSE 的表格資料轉換為與 `yfinance-go` 統一格式之資料欄位以供上游使用。
