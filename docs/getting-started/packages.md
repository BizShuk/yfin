# yfin 架構與套件指南 (Architecture and Packages Guide)

本教學文件將逐一介紹 `yfin` 專案中的所有套件 (package)。本專案採用模組化設計，將職責清晰劃分，確保系統的強健性 (robustness)、可觀測性 (observability) 與高精度的資料正規化。

---

## 系統架構關聯圖

以下為專案中各套件的依賴與呼叫關係圖：

```mermaid
flowchart TD
    EXT["外部消費者 (stock, data)"] -->|"import"| FAC["facade/"]
    FAC -->|"呼叫"| YC["Client (facade/client.go)"]

    YC -->|"原始 API"| YS["svc/yahoo"]
    YC -->|"爬網備援"| SC["svc/scrape"]
    YC -->|"正規化"| MD["model/"]
    YC -->|"台灣市場"| TW["svc/twse"]

    YS -->|"HTTP 傳輸"| HX["utils/httpx"]
    SC -->|"HTTP 傳輸"| HX
    HX -->|"指標 / 追蹤"| OB["utils/obsv"]

    SOAK["cmd/soak (壓測 binary)"] --> YS
    SOAK --> SC

    CMD["cmd/ (CLI composition root)"] -->|"讀取"| CFG["config/ (YAML)"]
    CMD -->|"只經 facade"| FAC
```

---

## 1. `facade` (對外契約層)

### 職責說明

`facade` 是專案對外的唯一契約邊界 (public contract boundary)，暴露反射免除 (reflection-free) 的純 Go 結構 (plain Go structs)，供外部專案（`stock`、`data` 等）直接 import 使用。

### 關鍵檔案

- `client.go`: 公開的 `Client` 結構與建構子 `NewClient` / `NewClientWithConfig`，實作核心 fetch 方法與備援 (fallback) 機制。
- `bars.go`: 轉換並包裝 K 線 (bar) 資料；定義 `Bar` 與 `BarBatch`。
- `quote.go`: 轉換並包裝即時報價 (quote) 資料。
- `company_info.go`: 轉換並包裝公司基本資訊 (company info)。
- `fundamentals.go`: 定義 `FundamentalsLine` 與 `FundamentalsSnapshot`。
- `market_data.go`: 定義市場資料聚合型別。
- `news.go`: 定義 `NewsItem`。
- `samples/`: 可執行的最小範例（`go run ./facade/samples/...`）。

### 套件設計與運作

- 外部套件在匯入本專案時，應以此套件作為邊界，避免直接存取 `svc/` 或 `utils/`。
- 將內部的定點十進位 `ScaledDecimal` 轉化為標準的 `float64` 浮點數，便於資料序列化 (serialization) 與一般數值計算。
- 整合 `svc/yahoo` 的 JSON API/XHR 與 `svc/scrape` 的顯式爬蟲 surface；呼叫端依需求選擇 `Fetch*` 或 `Scrape*`，不進行隱式 fallback。
- 所有取回的資料會經由 `model/` 的 `Normalize*` 函式正規化後，以統一格式回傳給呼叫端。

---

## 2. `svc/yahoo` (Yahoo 原始 API 服務)

### 職責說明

`svc/yahoo` 負責與 Yahoo Finance 原始 HTTP API 介面進行通訊，回傳未加工過的原始結構給上層處理。

### 關鍵檔案

- `client.go`: Yahoo client 與固定 endpoint origin。
- `auth.go`: 管理 Yahoo API 認證所需的 `Cookie` 與 `Crumb` 機制，會自動刷新並快取。
- `bars.go`, `quotes.go`, `fundamentals.go`: 對接並解析對應的 API 報文 (raw responses)。
- `timeseries.go`: annual income、balance sheet、cash flow 的 fundamentals-timeseries fetch/decode。
- `news.go`: `POST /xhr/ncp` tickerStream fetch/decode。

### 套件設計與運作

- 由於 Yahoo Finance 頻繁變更 Crumb 授權機制，此套件負責實作 Crumb 的抓取與快取重試邏輯。
- 利用 `utils/httpx` 提供連線，以取得最原始的、未加工過的 Yahoo 資料結構。

---

## 3. `svc/scrape` (網頁爬蟲服務)

### 職責說明

`svc/scrape` 是 HTML 爬網引擎 (scraping engine)，用於在 API 故障或需存取付費欄位時的替代解決方案。

### 關鍵檔案

- `client.go`: 網頁抓取的核心控制流程。
- `robots.go`: 解析與嚴格遵守遠端網站 `robots.txt` 爬蟲協定的模組。
- `financials.go`, `analysis.go`, `statistics.go`: 解析特定 HTML/JSON 網頁節點。
- `extract_news.go`: 新聞 JSON 擷取模組。

### 套件設計與運作

- 每次抓取前會先請求並快取 `robots.txt`，確保所有請求路徑符合合規性標準。
- 專門擷取分析師預測 (`ScrapeAnalysis`)、詳細財務報表 (`ScrapeFinancials`) 與即時新聞內容 (`ScrapeNews`)，彌補免費 API 的欄位缺憾。
- 所有 HTTP 呼叫委派給 `utils/httpx`。

---

## 4. `model/` (純資料與正規化層)

### 職責說明

`model/` 是專案最低層的純資料型別 + 正規化邏輯，**不 import `svc/*`、`facade/` 或 `cmd/`**。原 `svc/norm/` 已併入此套件。

### 關鍵檔案

- `normalize.go`: 9 個 `Normalize*` 函式（`NormalizeBars` / `NormalizedQuote` / `NormalizeFundamentals` / `NormalizeMarketData` / `NormalizeCompanyInfo` / `NormalizeHolders` / `NormalizeInsider`）。
- `security.go`: `Security` + `InferMIC` + `CreateSecurity` + `ExchangeToMIC`。
- `scaled.go`: `ScaledDecimal` 與 `FromScaledDecimal` 等定點十進位輔助。
- `time.go`: `ToUTCDayBoundaries` epoch→day 轉換。
- `fx.go`: `FXConverter` / `FXMeta` / `MockFXConverter`。
- `yahoo_raw.go`: 8 個 facade-aligned SDK DTO + Yahoo 原始 API 結構。
- `scrape_convert.go`: scrape DTO 直通 model（`Scrape*ToSnapshot` / `ScrapeNewsToItems`）。

### 套件設計與運作

- 處理來自 Yahoo JSON API 的複雜欄位（如 `regularMarketPrice` 與各種時間欄位）。
- 當 API 未提供 MIC 時，根據 symbol 尾碼（如 `.TW` 代表台灣證券交易所，MIC 為 `XTAI`）進行靜態推導與動態快取，確保識別資訊完整。
- 對外由 `facade.*` 型別別名重導出；外部消費者可直接 import `model/` 取得 SDK DTO，或透過 `facade/` 取得別名（兩者為相同型別）。

---

## 5. `svc/twse` (台灣證券交易所服務)

### 職責說明

`svc/twse` 為台灣市場專屬的證券交易所 (Taiwan Stock Exchange) 數據處理服務，封裝 23 個 TWSE Open Data 端點。

### 關鍵檔案

- `client.go`: TWSE 資料獲取用戶端。
- `parser.go`: 解析 TWSE 官方公佈的 BWIBBU (個股本益比、殖利率及股淨比) 與 T86 (三大法人買賣超) 等開放資料。

### 套件設計與運作

- 由於部分台灣市場指標（如法人籌碼面、本益比歷史明細）在 Yahoo Finance 較難完整取得，此服務直接請求 TWSE Open Data。
- 經過專屬解析器後，將 TWSE 的表格資料轉換為與 `yfin` 統一格式之資料欄位以供上游使用。
- 對應 CLI 入口為 `yfin twse` 子指令群（定義於 `cmd/twse/`）。

---

## 6. `utils/httpx` (彈性 HTTP 用戶端)

### 職責說明

`utils/httpx` 提供一個具備高彈性 (resilience) 的 HTTP 用戶端，是所有對外 HTTP 呼叫的唯一入口。

### 關鍵檔案

- `client.go`: 核心用戶端，整合權杖桶限流、重試與最終 outcome 分類。
- `caller.go`: 將相對或完整 URL 轉為 GET request，並回傳 response body 與 `Meta`。
- `circuit_breaker.go`: 實作 rolling-window 與 single-probe 狀態轉移。
- `circuit_registry.go`: 依 upstream authority 隔離斷路器狀態。
- `errors.go`: 定義連線、超時與限流之自訂錯誤型態。

### 套件設計與運作

- 採用單一共享的 `http.Client` 以最大化重複利用連線（基於 `Keep-Alive` 協議）。Session rotation 已完全移除以簡化狀態管理。
- 整合了 QPS 速率限制；當 API 回傳 `429` 或偵測到失敗率過高時，利用指數退避 (exponential backoff) 與隨機抖動 (jitter) 進行延遲重試。
- 當單一 upstream authority 在 active window 內累積足夠樣本且失敗率到達閾值，該 host 的斷路器進入 `open` (開啟) 狀態，直到冷卻重置時間結束。
- 透過 `utils/obsv` 上報延遲、成功率與限流命中數。

---

## 7. `utils/cache` (資料更新頻率快取)

### 職責說明

`utils/cache` 實作了專屬於 Yahoo Finance 資料抓取的更新頻率快取機制 (caching mechanism)。

### 關鍵檔案

- `refresh.go`: 定義不同資料維度的快取保留期 (retention period)。
- `tickerlist.go`: 處理本地端 Ticker 清單的讀取與載入。

### 套件設計與運作

- 依據資料屬性定義更新間隔：如 `daily` (每日)、`monthly` (每月) 與 `quarterly` (每季)。
- 提供 `REFRESH_MAP` 規則引擎，避免對不常變動的資料（如公司 Profile、Holder 結構）頻繁進行網路請求。

---

## 8. `utils/obsv` (可觀測性基礎設施)

### 職責說明

`utils/obsv` 負責整合系統的可觀測性 (observability) 架構，包含度量指標 (metrics) 與追蹤 (tracing)。

### 關鍵檔案

- `obsv.go`: 提供 OpenTelemetry 與追蹤監控的初始化方法。
- `metrics.go`: 包裝 Prometheus 註冊器與度量計數指標。

### 套件設計與運作

- 對 HTTP 用戶端的每次傳輸過程進行攔截，收集延遲時間、成功/失敗率、速率限制命中數。
- 將可觀測性資料匯出至 `inf` 監控後端（VictoriaMetrics `:8428` 與 Loki `:3100`）。
- 對應的 Grafana dashboard 與 runbook 位於 `monitoring/`。

---

## 9. `config/` (頂層 YAML 設定 loader)

### 職責說明

`config/` 是位於 repo 根目錄的頂層 YAML 設定 loader（**非 `internal/`**），提供 `Load` 與 schema 驗證，並將子型別拆入 `config/types/` 子套件。

### 關鍵檔案

- `types/config.go`: 頂層 `Config` struct。
- `types/loader.go`: YAML 載入 + 環境變數插值 + 驗證。
- `types/adapters.go`: HTTP / scrape / FX 等子設定型別。
- `types/loader_test.go`: 解析測試。
- `effective.yaml`: 經由環境變數插值後產生的實際生效設定。不同環境以 `app.env` 區分。

### 套件設計與運作

- 在系統啟動時載入組態檔案，檢查並行工作線數 (concurrency workers)、QPS 限流速率與斷路器參數之合理性。
- 將設定結果傳遞給 `utils/httpx` 作為初始化基礎。
- 應用設定目錄固定為 `~/.config/yfin/`（由 `gosdk` 提供，慣例不提供自訂選項）。

---

## 10. `cmd/` (CLI composition root)

### 職責說明

`cmd/` 是命令列介面的組合根 (composition root)，採用子套件 (sub-package) 結構，每個子指令群自治於自己的子目錄。

### 關鍵檔案（cmd/ 根層 helpers）

- `root.go`: 根指令與 global flags。
- `client.go`: 共用的 `Client` 初始化器。
- `global.go`: 全域旗標與 logger 初始化。
- `build.go`: build-time 變數（版本、commit）。
- `exitcodes.go`: 退出碼常數與測試。
- `fetch.go`: 共用的 fetch helper。

### 子套件（每個獨立可建置）

- `cmd/admin/`: `config-effective`、`version` 等管理子指令。
- `cmd/dispatch/`: 批次調度（無對外 `yfin dispatch` 指令，作為 batch 的內部封裝）。
- `cmd/fundamentals/`: `fundamentals`、`comprehensive-stats`、`comprehensive-profile`。
- `cmd/market/`: `pull`、`quote`。
- `cmd/scrape/`: `scrape`（4 個互斥 mode）。
- `cmd/twse/`: `twse`（23 個端點）。
- `cmd/soak/`: 獨立 binary（見第 11 節）。
- `cmd/samples/`: 可執行的 CLI 子指令範例。
- `cmd/tools/`: 輔助工具。

### 套件設計與運作

- 使用 `cobra` 套件管理命令列參數；每個子套件對外暴露 `Register(parentCmd)` 函式，由 `main.go` 註冊。
- 解析 `config/` 的 YAML 設定，初始化連線池，並根據輸入指令執行對應的抓取、發佈或壓測。

---

## 11. `cmd/soak` (壓力測試獨立 binary)

### 職責說明

`cmd/soak` 負責長時間穩定性測試 (soak testing) 的調度與執行，是獨立的 CLI binary（`go run ./cmd/soak`）。

### 關鍵檔案

- `orchestrator.go`: 壓測工作的協調器。
- `worker.go`: 並行拉取工作單元。
- `probes.go`: 數值與邏輯正確性探針。
- `memory.go`: 定期調用 runtime 進行記憶體使用增長與洩漏分析。

### 套件設計與運作

- 在開發與持續整合 (CI) 階段，被用來進行高強度壓測，長時間以並行 goroutine 拉取真實市場資料。
- 正確性探針會對產出的 `NormalizedBar` 進行極限值、時間順序性等比對，並產出壓測報告。
- 直接呼叫 `svc/yahoo` 與 `svc/scrape`，繞過 `facade` 以貼近底層行為。

---

## 12. `main.go` (入口點)

### 職責說明

`main.go` 位於 repo 根目錄，是 `yfin` CLI 的真正進入點；它是最薄的殼，轉發 6 個 `Register` 呼叫後呼叫 `cmd.Execute()`。

### 套件設計與運作

- 依序註冊 `cmd/admin`、`cmd/dispatch`、`cmd/fundamentals`、`cmd/market`、`cmd/scrape`、`cmd/twse` 至根指令。
- 設定 logger（slog，來自 `gosdk/log`）與 metrics exporter，再將控制權交給 cobra。
- 任何新增的子指令群都應新增一條 `Register` 呼叫，並對應建立 `cmd/<name>/` 子套件；不要直接在 `main.go` 撰寫邏輯。
