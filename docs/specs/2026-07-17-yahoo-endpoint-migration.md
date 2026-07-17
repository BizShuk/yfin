# Yahoo Timeseries and News Endpoint Migration

## 結論

將 `income`、`balance`、`cashflow` 從 Yahoo HTML 頁面改接
`fundamentals-timeseries`，並將 `news` 改接 Yahoo Finance `xhr/ncp`。
這四個 endpoint 已確認可由 Go 標準 `net/http` 存取，不需要新增 browser impersonation
或 `curl_cffi` 類型的依賴。

遷移保持現有 CLI JSON schema：財報仍回傳單一 `model.FundamentalsSnapshot`，新聞仍回傳
`[]model.NewsItem`。其餘需要 cookie、crumb 或瀏覽器指紋的 22 個失敗不在本次範圍。

## 背景與根因

AAPL 30-command live batch 在修復 host-scoped circuit breaker 後得到
`4 success / 26 failed / 0 circuit-open`。其中：

- `income`、`balance`、`cashflow` 與 `news` 因 Yahoo HTML 頁面回傳 `503` 失敗。
- Python `yfinance` 的三個財報 surface 已使用
  `/ws/fundamentals-timeseries/v1/finance/timeseries/{symbol}`，不是解析 HTML。
- Python `yfinance` 的新聞 surface 已使用 `POST /xhr/ncp`，不是解析 quote news HTML。
- 直接 live probe 確認上述兩個 endpoint family 以標準 Go/Python HTTP transport 可回傳 `200`。

因此這四個失敗是 endpoint family 過時，不是 retry、breaker 或資料模型問題。

## 範圍

本次包含：

- 新增 Yahoo annual fundamentals-timeseries fetch/decode。
- 支援 income、balance sheet、cash flow 三組既有 model key。
- 新增 Yahoo `latestNews` XHR POST fetch/decode。
- 使用可 replay 的 stdlib request body，並以回歸測試確認 retry 內容一致。
- facade 暴露 model-only 方法，dispatch 的四個 command 改走新方法。
- unit、race、build、分層與 live batch 驗證。

本次不包含：

- quoteSummary、v7 quote 或 options 的 cookie/crumb/browser fingerprint 修復。
- earnings/analysis HTML scrape 遷移。
- 新增第三方 HTTP/TLS 套件。
- 修改 CLI command manifest 或輸出 schema。
- 移除仍供其他 API 使用的既有 `Scrape*` 方法。

## 架構

```text
cmd/dispatch
└── facade.Client
    ├── FetchIncomeStatement
    ├── FetchBalanceSheet
    ├── FetchCashFlowStatement
    └── FetchNews
        └── svc/yahoo.Client
            ├── FetchFinancialStatement
            └── FetchNews
                └── model.FundamentalsSnapshot / model.NewsItem
```

`cmd/*` 不 import `svc/*`。`facade` 對外簽名只包含 facade alias 或 `model` 型別，不洩漏
`svc/yahoo` 的 endpoint envelope。

## Fundamentals-timeseries 契約

請求使用絕對 URL：

```text
GET https://query2.finance.yahoo.com/ws/fundamentals-timeseries/v1/finance/timeseries/{symbol}
    ?symbol={symbol}
    &type=annual<Type1>,annual<Type2>,...
    &period1=1483142400
    &period2=<next UTC day>
```

`period1` 與 Python yfinance 對齊為 2016-12-31 UTC；Yahoo 目前最多回傳四個年度。
每個 series 只取 `asOfDate` 最新且含 `reportedValue.raw` 的 annual point。不同欄位允許缺值；
缺少的欄位不產生零值 line。若所有欄位都缺失則回傳明確錯誤。

時間欄位：

- `PeriodEnd`：Yahoo `asOfDate` 的 UTC 00:00。
- `PeriodStart`：`PeriodEnd.AddDate(-1, 0, 1)`，表示包含首尾的年度期間。
- `AsOf`：所有成功 line 中最大的 `PeriodEnd`。

`Source` 使用 `yahoo/fundamentals-timeseries/<statement>`。`svc/yahoo` 填入 symbol、source、
as-of 與 lines；`facade` 使用既有 MIC inference cache 補上 `MIC`。

### 欄位映射

Income：

| Yahoo annual type | Model key |
| --- | --- |
| `TotalRevenue` | `total_revenue` |
| `CostOfRevenue` | `cost_of_revenue` |
| `GrossProfit` | `gross_profit` |
| `OperatingExpense` | `operating_expense` |
| `OperatingIncome` | `operating_income` |
| `NetNonOperatingInterestIncomeExpense` | `net_non_operating_interest_income_expense` |
| `OtherIncomeExpense` | `other_income_expense` |
| `PretaxIncome` | `pretax_income` |
| `TaxProvision` | `tax_provision` |
| `NetIncomeCommonStockholders` | `net_income` |
| `BasicEPS` | `eps_basic` |
| `DilutedEPS` | `eps_diluted` |
| `BasicAverageShares` | `shares_outstanding_basic` |
| `DilutedAverageShares` | `shares_outstanding_diluted` |
| `TotalExpenses` | `total_expenses` |
| `NormalizedIncome` | `normalized_income` |
| `EBIT` | `ebit` |
| `EBITDA` | `ebitda` |
| `ReconciledCostOfRevenue` | `reconciled_cost_of_revenue` |
| `ReconciledDepreciation` | `reconciled_depreciation` |
| `NormalizedEBITDA` | `normalized_ebitda` |

Balance sheet：

| Yahoo annual type | Model key |
| --- | --- |
| `TotalAssets` | `total_assets` |
| `TotalDebt` | `total_debt` |
| `CommonStockEquity` | `shareholders_equity` |
| `WorkingCapital` | `working_capital` |
| `TangibleBookValue` | `tangible_book_value` |

Cash flow：

| Yahoo annual type | Model key |
| --- | --- |
| `OperatingCashFlow` | `operating_cash_flow` |
| `InvestingCashFlow` | `investing_cash_flow` |
| `FinancingCashFlow` | `financing_cash_flow` |
| `FreeCashFlow` | `free_cash_flow` |
| `CapitalExpenditure` | `capital_expenditure` |

## News XHR 契約

請求：

```text
POST https://finance.yahoo.com/xhr/ncp?queryRef=latestNews&serviceKey=ncp_fin
Content-Type: application/json

{"serviceConfig":{"snippetCount":10,"s":["AAPL"]}}
```

解析 `data.tickerStream.stream`：

- 忽略含 `ad` 資料的項目。
- `Title` 取 `content.title`。
- `URL` 優先取 `content.canonicalUrl.url`，否則取 `content.clickThroughUrl.url`。
- `Source` 取 `content.provider.displayName`。
- `Summary` 取 `content.summary`。
- `PublishedAt` 以 RFC3339 解析 `content.pubDate`。
- `Symbols` 固定包含請求 symbol。
- 缺少 title 或 URL 的項目不輸出。

## POST retry body

News request 必須以 `http.NewRequestWithContext(..., bytes.NewReader(payload))` 建立。stdlib 會為
`bytes.Reader` 設定 `Request.GetBody`；目前 Go transport 與既有 `httpx.Client.Do` 組合在 outer
retry 時能重送相同內容。

實作前的回歸測試已確認第一次 `500` 後，第二次 attempt 收到完全相同的 JSON。因此本次不修改
`httpx.Client.Do`，避免在現有行為已滿足需求時加入額外 body lifecycle 邏輯。測試保留作為
news POST 的 transport contract。

## 錯誤處理

- HTTP status 仍由 `httpx.Client` 統一處理。
- JSON malformed、Yahoo response error、空 timeseries、所有 mapped series 缺值均回傳帶 context 的 error。
- 單一 news item 的日期 malformed 不使整批失敗；該 item 的 `PublishedAt` 保持 zero time。
- XHR response 沒有有效新聞時回傳空 slice，不視為 transport failure。

## 驗收

Deterministic gate：

- `go test ./... -count=1`
- `go test -race ./utils/httpx ./svc/yahoo ./facade ./cmd/dispatch -count=1`
- `go build -o /tmp/yfin-endpoint-migration-check .`
- `cmd/*` 對 `svc/*` import 檢查無輸出。
- `git diff --check`

Live gate：

- 執行 AAPL 固定 30-command batch。
- `income`、`balance`、`cashflow`、`news` 成功產生 artifact。
- 批次至少 `8 success / 22 failed`。
- 不出現 `circuit breaker is open`。
