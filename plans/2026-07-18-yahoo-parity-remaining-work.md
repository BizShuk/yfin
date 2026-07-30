# Yahoo 30-command Parity 剩餘工作實作計畫 (Implementation Plan)

> 給 Agentic Workers：請使用 `test-driven-development` 並逐一任務執行本計畫。各步驟使用核取方塊 (`- [ ]`) 語法進行追蹤。

`Status: Active`

`Goal:` 確保 Batch Artifacts 的真實性，將剩餘可避免的 HTML 路徑替換為穩定的 Yahoo JSON Surface，並讓 `AAPL` 與 `2330.TW` 皆通過固定的 30-command Semantic Parity Gate，同時不誇大上游 Anti-bot 的限制。

`Architecture:` 維持 `cmd → facade → svc → model` 架構。Batch 擁有 Artifact Lifecycle；`svc/yahoo` 擁有 HTTP Request 與 Decode 契約；Facade 僅暴露 Model/Facade 型別；Parity 腳本維持作為外部 Acceptance Oracle。不引進任何 Production Python、`curl_cffi`、TLS 偽裝函式庫或 Browser Runtime。

`Current live baseline:` 2026-07-17 記錄的最後一次正式運行結果為 AAPL 的 `8 success / 22 failed`。八個正常的 Command 為 `history`、`actions`、`income`、`balance`、`cashflow`、`news`、`isin` 與 `metadata`。其餘 22 個結果包含 13 個 quoteSummary/auth-family 失敗、7 個 analysis/analyst HTML `503` 失敗、`options` `401` 以及無資料列的 `earnings-dates`。這是歷史數據，在宣告當前 Parity 之前必須先重新整理。

`Tech stack:` Go 1.26.0, Cobra, stdlib HTTP/JSON/filesystem 套件, 現有 `utils/httpx`, Python yfinance oracle 腳本, Bash, 與 `testify`。

## 全局限制 (Global Constraints)

- 請勿引進 Production Python、`curl_cffi`、TLS 偽裝、Playwright 或新的 HTTP Transport 依賴。
- 保持 `facade.Client` 為 `cmd/dispatch` 所使用的唯一 Fetch Handle。
- 保持 `model` 為最底層 (Lowest Layer)。
- 保持 Runtime Artifacts 儲存在 `~/.config/yfin/data/raw/` 下。
- 保留固定的 30-command Manifest 與特定的 JSON 語義。
- 將 Deterministic Tests 與 Live Yahoo Parity 視為獨立的 Gate。
- 只有當當前 JSON Artifact 存在且過期的 `_failed` Artifact 不存在時，Live Command 才算成功。
- 在 AAPL Gate 回傳 0 之前，請勿執行 `2330.TW` Acceptance。
- 如果 Yahoo 對某個 Command 仍需要 Browser 等價的 Fingerprint，請將其記錄為明確的 Blocker；切勿放寬 Gate 或將其標記為完成。

## 失敗清單 (Failure Inventory)

| Family | Commands | 上次觀察到的失敗 (Last observed failure) |
| --- | --- | --- |
| quoteSummary/auth | `info`, 三個 holder commands, 三個 insider commands, `recommendations`, `recommendations-summary`, `upgrades`, `calendar`, `sec-filings`, `sustainability` | crumb `429` 或 auth-family circuit-open |
| analysis JSON candidates | `earnings-history`, `eps-trend`, `eps-revisions`, `earnings-estimates`, `revenue-estimates`, `growth-estimates`, `price-targets` | Yahoo HTML `503` |
| options | `options` | `401 Unauthorized` |
| earnings calendar | `earnings-dates` | HTML Parser 回傳無資料列 (no rows) |

## 檔案結構 (File Structure)

- 修改 `cmd/dispatch/batch.go`：在 Non-failure Terminal States 後清除過期的 Error Artifacts。
- 修改 `cmd/dispatch/batch_test.go` 與 `batch_integration_test.go`：Artifact-lifecycle 測試。
- 修改 `scripts/verify-yf-parity.sh`：即使 Go 離開時非零，也總是執行 Classification 與 Comparison。
- 建立 `skills/scripts/classify_yfin_failures.py`：Current-run Family Summary。
- 建立 `skills/scripts/test_classify_yfin_failures.py`：Deterministic Classifier 測試。
- 建立 `svc/yahoo/analysis_quotesummary.go` 與測試：quoteSummary Analysis 模組與 Projections。
- 修改 `facade/client_yahoo.go` 與測試：Plain Analysis Surfaces。
- 修改 `cmd/dispatch/dispatch.go` 與測試：切換 (Cut over) 七個 Analysis Commands。
- 建立 `svc/yahoo/earnings_calendar.go` 與測試：Visualization JSON Probe/Decoder，在驗證 Freshness 前與 Active HTML Path 保持分離。
- 修改 `svc/yahoo/options.go` 與測試：若 Live Probe 證明有效，則更新 Authenticated Option-chain Request 契約。
- 更新 `plans/2026-07-18-yahoo-parity-remaining-work.md`：在每次正式運行後附加確切的 Live Evidence。
- 僅在行為變更後更新 `README.md`、`CLAUDE.md`、`docs/cli/commands.md` 及 `docs/operations/error-handling.md`。

---

### Task 1: 真實的 `_failed` Artifact Lifecycle

`Files:`

- 修改 `cmd/dispatch/batch.go`。
- 修改 `cmd/dispatch/batch_test.go`。
- 修改 `cmd/dispatch/batch_integration_test.go`。

`Interfaces:`

- 產出 `func clearCommandFailure(rawDir, ticker, command string) error`。
- `success`、`skipped` 或 `not_found` 的結果會移除 `_failed/<ticker>.<command>.err`。
- `os.ErrNotExist` 視為成功；其他任何移除錯誤皆會顯現於 `tickerResult.Errors` 中。

- [ ] `Step 1: 撰寫 Failing Stale-error 測試`

預先建立 `_failed/AAPL.good.err`，執行成功的 Fetch，並斷言 (Assert) JSON Artifact 存在而過期的 Error 不存在。針對 `skipped` 與 `not_found` 重複此流程。對於移除錯誤的情境，將 Error Path 替換為非空目錄，使 `os.Remove` 在所有支援的平台上皆確定性地回傳錯誤。

- [ ] `Step 2: 執行單獨測試並驗證 RED`

```bash
go test ./cmd/dispatch -run 'TestRunBatchForTicker_(ClearsStaleFailure|SkipClearsStaleFailure|NotFoundClearsStaleFailure|FailureCleanupError)' -count=1
```

預期結果：Stale Error 斷言失敗，因為 `runBatchForTicker` 從未移除它們。

- [ ] `Step 3: 在 Terminal-state 邊界實作清理`

僅在成功的 JSON 重命名後、確認 Cache Skip 後，或在當前 Not-found 分類後呼叫 `clearCommandFailure`。請勿在 Fetch 開始前清理 Error；被中斷的 Request 不得擦除最後已知的 Failure。清理錯誤會將 Command 狀態改為 `failed` 並記錄於 `tickerResult.Errors` 中。

- [ ] `Step 4: 執行 Dispatch 與 Lifecycle 整合測試`

```bash
go test ./cmd/dispatch -count=1
```

預期結果：所有測試通過，包含現有的 Failure-artifact 行為。

### Task 2: Current-run Parity 分類 (Classification)

`Files:`

- 建立 `skills/scripts/classify_yfin_failures.py`。
- 建立 `skills/scripts/test_classify_yfin_failures.py`。
- 修改 `scripts/verify-yf-parity.sh`。

`Interfaces:`

- Classifier 讀取 `--go-root`、`--ticker`、`--day` 以及來自 `skills/scripts/config.py` 的固定 Commands。
- Classifier 印出確切的 `success`、`failed`、`missing` 及 Family 計數，且除非所有 30 個 Commands 皆有合法的 JSON 且無對應的 Error，否則離開時返回非零值。
- Shell Runner 分別擷取 Python、Go、Classifier 與 Semantic-comparator 狀態，若有任何階段失敗則回傳非零值。

- [ ] `Step 1: 撰寫 Classifier Fixture 測試`

為記錄的 `8/22` 分拆、Stale-error-plus-success 衝突、無效 JSON 及完整的 `30/30` 建立臨時 Fixtures。斷言穩定的 Family Counts 與 Exit Codes。

- [ ] `Step 2: 執行測試並驗證 RED`

```bash
python3 skills/scripts/test_classify_yfin_failures.py
```

預期結果：因為 Classifier 還不存在而 Import 失敗。

- [ ] `Step 3: 實作 Classifier`

依本計畫中的四個固定 Family 進行分類。Success Artifact 加上 `_failed` 檔案視為 `conflict` 而非 Success。僅讀取請求的 UTC 日期。

- [ ] `Step 4: 使 Shell Gate 執行每個診斷階段`

僅在 Go Batch 與 Comparison Commands 周圍暫時停用立即退出 (`set -e`)，擷取狀態，恢復 `set -e`，印出四個階段代碼，若任一非零則離開並回傳 `1`。

- [ ] `Step 5: 執行 Deterministic 腳本測試`

```bash
python3 skills/scripts/test_classify_yfin_failures.py
bash -n scripts/verify-yf-parity.sh
```

預期結果：兩者皆通過。

### Task 3: 替換七個 Analysis HTML Commands

`Files:`

- 建立 `svc/yahoo/analysis_quotesummary.go`。
- 建立 `svc/yahoo/analysis_quotesummary_test.go`。
- 修改 `facade/client_yahoo.go`。
- 修改 `facade/client_yahoo_test.go`。
- 修改 `cmd/dispatch/dispatch.go`。
- 修改 `cmd/dispatch/dispatch_test.go`。

`Interfaces:`

- `earnings-history` 解碼 quoteSummary 模組 `earningsHistory`。
- `earnings-estimates`、`revenue-estimates`、`eps-trend` 與 `eps-revisions` 從 `earningsTrend` 投影 (Project) 其對應的巢狀欄位。
- `growth-estimates` 將來自 `earningsTrend` 的個股成長欄位與 `industryTrend`、`sectorTrend` 及 `indexTrend` 的估算結合。
- `price-targets` 投影 `financialData` 的 Target-price 欄位。
- Facade 方法僅回傳 `model` 或 Facade 自有的 DTOs。

- [ ] `Step 1: 新增基於 Fixture 的 Decoder 與 Projection 測試`

涵蓋遺失的模組、明確的零值 (Zero values)、可為 Null 的 Raw 欄位、確定性的資料列順序 (Row order)，以及跨上游模組共享的六個 Analysis 維度的獨立 Payload。

- [ ] `Step 2: 執行 Service 測試並驗證 RED`

```bash
go test ./svc/yahoo -run 'TestDecode(Analysis|EarningsTrend|PriceTargets)' -count=1
```

預期結果：因為 Decoder 還不存在而 Build 失敗。

- [ ] `Step 3: 實作 Request 與 Decode 契約`

使用現有的 quoteSummary Request 構建、Crumb 處理與 `yahoo-auth` Circuit Group。請勿解析 HTML，也不要在 Service 邊界跨界重用 Scrape DTOs。

- [ ] `Step 4: 新增 Facade 方法並進行 Dispatch 切換 (Cutover)`

僅搬移列出的七個 Registry 項目。新增一個 Source-contract 測試，拒絕在 `cmd/dispatch/dispatch.go` 中使用 `ScrapeAnalysisDimension` 與 `ScrapeAnalystInsights`。

- [ ] `Step 5: 執行單獨與完整的 Deterministic 測試`

```bash
go test ./svc/yahoo ./facade ./cmd/dispatch -count=1
go test ./... -count=1
```

預期結果：所有測試通過。在 Task 6 之前不宣稱 Live 成功。

### Task 4: Earnings-calendar 能力 Gate

`Files:`

- 建立 `svc/yahoo/earnings_calendar.go`。
- 建立 `svc/yahoo/earnings_calendar_test.go`。
- 建立 `scripts/probe-yahoo-earnings-calendar.sh`。
- 僅在 Live Freshness Gate 通過時，修改 `svc/yahoo/earnings_dates.go`、`facade/yahoo.go` 與 `cmd/dispatch/dispatch_test.go`。

`Interfaces:`

- 為 Yahoo 的 `POST /v1/finance/visualization` JSON Surface 新增一個並行的 (Side-by-side) Decoder，使用 `entityIdType=earnings`、Ticker 相等性過濾器、`startdatetime` 降序排序，以及六個 yfinance 欄位。
- 當前 Python yfinance 源碼指出此 Surface 已於 2025 年停止更新；除非 AAPL 與 2330.TW 的 Freshness 在同一次運行中匹配 Python Oracle，否則無法替換 HTML。
- POST Request Bodies 必須在 Retries 期間保持可重放 (Replayable)。

- [ ] `Step 1: 撰寫 Request-contract 與 Decoder 測試`

斷言 Method、Endpoint、確切的 JSON Filter/Sort/Fields、Content Type、Response Mapping、Empty Result、Nullable EPS 欄位以及 Retry Body Replay。

- [ ] `Step 2: 執行單獨測試並驗證 RED`

```bash
go test ./svc/yahoo -run 'Test(Fetch|Decode)EarningsCalendar' -count=1
```

預期結果：因為 Visualization Client 還不存在而 Build 失敗。

- [ ] `Step 3: 實作隔離的 Visualization Probe`

使用現有的 `httpx.Client.Do` Transport 與 `yahoo-web` Group。在此檢查點保持 Production `FetchEarningsDates` 路徑不變。

- [ ] `Step 4: 在切換前比較 Live Freshness`

Probe 印出來自 Visualization 與 Python Oracle 的最新與最舊 Event Timestamps（針對兩個 Tickers），且不印出 Cookies。僅當兩個資料集涵蓋相同的當前 Event Window 時才進行切換。若 Visualization 過期或 HTML 仍被 Fingerprint 封鎖，將 `earnings-dates` 記錄為 Blocked，且不進行 Production 切換。

- [ ] `Step 5: 執行單獨與完整的 Yahoo 測試`

```bash
go test ./svc/yahoo ./facade ./cmd/dispatch -count=1
```

預期結果：所有測試通過。

### Task 5: Auth 與 Options 能力 Gate

`Files:`

- 建立 `scripts/probe-yahoo-auth.sh`。
- 建立包含 `//go:build integration` 的 `svc/yahoo/auth_live_test.go`。
- 僅在 Authenticated stdlib Request 成功時，修改 `svc/yahoo/options.go` 與測試。
- 將 Probe Evidence 附加至此計畫中。

`Interfaces:`

- 使用與 Production 相同的已設定 stdlib Transport 探測 `quoteSummary`、`options` 以及 Crumb/Cookie Bootstrap。
- 記錄 HTTP Status、Authority、Family，以及 Python yfinance 在相同 Ticker 與時間視窗下是否成功。
- 切勿印出 Cookie 或 Crumb 數值。

- [ ] `Step 1: 新增選擇性 (Opt-in) 整合 Probe`

需要 `YFIN_LIVE_YAHOO=1`；否則跳過。以 Retry Max `1` 順序測試 AAPL，使 Circuit Breaker 放大不會隱藏根狀態。

- [ ] `Step 2: 執行正式的 stdlib Probe`

```bash
YFIN_LIVE_YAHOO=1 go test -tags=integration ./svc/yahoo -run TestLiveYahooAuthCapability -count=1 -v
```

預期結果：印出確切的 Statuses 且不洩漏 Secrets。

- [ ] `Step 3: 僅套用有證據支持的 Options 變更`

若相同的 Cookie/Crumb 契約使 Options 回傳 `2xx`，則新增它及 Fixture 測試，並維持 `yahoo-options` 隔離。若其在 Python 成功的情況下仍為 `401`，將 `options` 記錄為 Fingerprint-blocked，且不進行推測性的 Production 變更。

- [ ] `Step 4: 記錄 quoteSummary 邊界`

若 stdlib quoteSummary 在 Python yfinance 成功時仍保持 `429`，請保留所有 13 個 Auth Commands 為明確 Blocked。解決該狀況需要單獨獲批的 Transport-policy 變更，因為本計畫禁止 Fingerprint Impersonation 依賴。

### Task 6: 正式 AAPL 隨後 2330.TW 驗收 (Acceptance)

`Files:`

- 修改 `README.md`、`CLAUDE.md`、`docs/cli/commands.md` 與 `docs/operations/error-handling.md`。
- 將確切結果附加至此計畫中。

- [ ] `Step 1: 執行 Deterministic Gates`

```bash
go test ./... -count=1
go test -race ./... -count=1
go build ./...
go vet ./...
go list -f '{{.ImportPath}} {{join .Imports " "}}' ./cmd/... | grep 'github.com/bizshuk/yfin/svc/'
git diff --check
```

預期結果：Test, race, build, vet, 及 diff 檢查通過；Import Grep 不輸出任何內容。

- [ ] `Step 2: 執行 AAPL Parity`

```bash
./scripts/verify-yf-parity.sh AAPL
```

預期完成結果：`30 success`，零 failed/missing/conflict，Semantic Comparator 離開時返回 `0`。

- [ ] `Step 3: 僅在 AAPL 通過後執行 2330.TW Parity`

```bash
./scripts/verify-yf-parity.sh 2330.TW
```

預期完成結果：`30 success`，零 failed/missing/conflict，Semantic Comparator 離開時返回 `0`。

- [ ] `Step 4: 同步 Active Documentation`

僅記錄兩個運行皆證明的 Endpoints 與 Live Support。若 Task 5 仍處於 Fingerprint-blocked 狀態，保持此計畫為 Active 狀態且不宣稱完整 Parity。

## 最終驗收標準 (Final Acceptance Criteria)

- [ ] 當前 success、skipped 及 not-found 的結果不能與過期的 Error Artifacts 共存。
- [ ] Runner 總是回報所有階段的狀態。
- [ ] 七個 Analysis Commands 不再使用 Yahoo HTML Scrape 路徑。
- [ ] Earnings Dates 使用 JSON 且一致地處理空結果。
- [ ] Options 行為有證據支持。
- [ ] AAPL 30-command Semantic Parity 離開時返回 `0`。
- [ ] 2330.TW 30-command Semantic Parity 離開時返回 `0`。
- [ ] Deterministic Gates 通過且 `cmd → svc` 直接 Import 為零。
