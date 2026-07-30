# 錯誤處理與疑難排解 (Error Handling and Troubleshooting)

本指南說明 `yfin` 目前的 exit codes、HTTP / scrape error 分類，以及可直接執行的診斷流程。CLI surface 以 `yfin --help` 與各 subcommand 的 `--help` 為準。

## Exit Codes

`cmd/exitcodes.go` 定義四個穩定值：

| Code | Constant | Meaning |
| --- | --- | --- |
| `0` | `ExitSuccess` | 命令正常完成 |
| `1` | `ExitGeneral` | Network、provider、parse 或其他 runtime failure |
| `2` | `ExitPaidFeature` | Paid-only Yahoo endpoint 未授權 |
| `3` | `ExitConfigError` | Flag、YAML、validation 或 observability initialization 錯誤 |

`fundamentals` 遇到 paid-subscription / 401 類錯誤時使用 code `2`。設定檔無法載入、缺少必要 flag 或 config validation 失敗時使用 code `3`。

## 第一層診斷 (First Response)

先確認 binary、effective config 與目標 command contract：

```bash
yfin version
yfin config --print-effective
yfin scrape --help
```

使用自訂設定檔時，`--config` 是 Global Flag：

```bash
yfin --config ./config/production.yaml config --print-effective
```

未指定 `--config` 時會載入 `config/effective.yaml`。Effective config output 會 redact secret references；一般欄位如 `scrape.endpoints.key_statistics` 必須保留原值。

## 常見 CLI Validation Errors

| Trigger | Error shape |
| --- | --- |
| `pull` 同時缺少或同時提供 `--ticker` / `--universe-file` | `either --ticker or --universe-file must be specified` |
| `pull` 缺少 `--start` / `--end` | `--start and --end are required` |
| `--adjusted` 非 `raw` / `split_dividend` | `--adjusted must be 'raw' or 'split_dividend'` |
| `--out` 非 `json` / `parquet` | `--out must be 'json' or 'parquet'` |
| 未指定或同時指定多個 scrape modes | `exactly one of --check, --preview-json, or --preview-news is required` |
| `--check` 缺少 `--endpoint` | `--endpoint is required for --check mode` |
| `--preview-json` 缺少 `--endpoints` | `--endpoints is required for --preview-json mode` |
| YAML schema / invariant 不合法 | `Failed to load configuration: ...` |

`pull` 與 `quote` 沒有 `--preview` flag；未要求 local export 時，它們仍會將摘要寫到 stdout。`--out parquet` 目前會回報尚未實作。

## Scrape 診斷模式

`scrape` 提供三種 mode：

| Mode | Required flags | Purpose |
| --- | --- | --- |
| `--check` | `--ticker X --endpoint Y` | Fetch response 並檢查 connectivity，不解析 DTO |
| `--preview-json` | `--ticker X --endpoints a,b` | 執行 extractor 並輸出 JSON |
| `--preview-news` | `--ticker X` | 執行 news parser |

標準流程：

```bash
# 1. Connectivity
yfin scrape --check --ticker AAPL --endpoint key-statistics

# 2. Extractor
yfin scrape --preview-json --ticker AAPL --endpoints key-statistics

# 3. Debug logging
yfin --log-level debug scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics

# 4. Cross-symbol verification
for ticker in AAPL MSFT GOOGL; do
  yfin scrape --check --ticker "$ticker" --endpoint key-statistics
done
```

`--force` 是無額外效果的 compatibility flag，不會停用 robots.txt enforcement。Robots behavior 由 `scrape.robots_policy` 控制。

## HTTP Error Classification

`utils/httpx/errors.go` 提供 typed errors 與 classifiers：

- Retryable：HTTP `429`, `500`, `502`, `503`, `504`，以及 `ErrTooManyRequests`、`ErrServerUnavailable`、`ErrTimeout`。
- Fatal：HTTP `400`, `401`, `403`, `404`, `422`，以及 `ErrClientConfig`、`ErrDecode`。
- `ErrCircuitOpen` 表示該 authority + logical group 的 circuit breaker 已開啟；應降低 request pressure 並等待 reset。

HTTP client 對一個 logical request 的所有 retries 只記錄一次最終 outcome。Session rotation 已移除；`--sessions` 僅為 compatibility flag，沒有 runtime effect。

### Network / Timeout

常見訊息包括 `context deadline exceeded`、`connection refused`、`no such host`。先檢查 DNS / network，再以 Global Flag 調整 timeout：

```bash
yfin --timeout 30s scrape --check --ticker AAPL --endpoint profile
```

### Rate Limiting

HTTP `429` 或反覆 `503` 時，單次 scrape 先降低 QPS：

```bash
yfin --qps 0.5 scrape \
  --check \
  --ticker AAPL \
  --endpoint key-statistics
```

Yahoo API 的 YAML knobs 位於 `rate_limit` / `retry` / `circuit_breaker`。Scrape-specific HTTP defaults 屬於 Go `httpx.Config`，不是 `scrape.qps` 或 `scrape.retry` YAML 欄位。
`pull` universe / multi-ticker `quote` 才有 worker pool；這兩種 workload 可再降低 `--concurrency`。

## Scrape Error Classification

`svc/scrape/errors.go` 定義 `ScrapeError`：

| Error | Meaning |
| --- | --- |
| `ErrRobotsDenied` | robots.txt 不允許該 path |
| `ErrTimeout` | Request timeout |
| `ErrRateLimited` | Provider rate limit |
| `ErrCircuitOpen` | Circuit breaker open |
| `ErrContentTooLarge` | Response 超過 body cap |
| `ErrNoQuoteSummary` | 找不到 quote-summary payload |
| `ErrJSONUnescape` / `ErrJSONDecode` | Embedded JSON 處理失敗 |
| `ErrMissingFieldBase` / `ErrSchemaDriftBase` | Extractor schema drift |
| `ErrNewsNoArticles` / `ErrNewsParse` | News 缺資料或 parsing failure |

遇到 parsing drift 時，先保存 debug log，再以多個 symbols 重現：

```bash
for ticker in AAPL MSFT BRK-B; do
  yfin --log-level debug scrape \
    --preview-json \
    --ticker "$ticker" \
    --endpoints key-statistics
done
```

修正應落在對應的 `svc/scrape` extractor，並新增 fixture / parser test；不要在 CLI layer 補 provider-specific parsing。

## Paid Fundamentals

`FetchFundamentalsQuarterly` 可能需要 Yahoo paid subscription。無授權時可明確選擇 scrape surface，而不是隱式 fallback：

```bash
yfin scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints financials,key-statistics,analysis,profile
```

Annual income statement、balance sheet 與 cash flow 另有免費 fundamentals-timeseries methods；請依所需資料 surface 選擇。

## Empty Data Semantics

結構有效但沒有 observations 的 chart response 不等於 transport error。Downstream 應檢查 `len(batch.Bars)`，並保留 symbol / MIC / provider metadata。不要自行補造 currency 或 bars。

對 `Normalized*` 的 `ScaledDecimal` 使用 `model.FromScaledDecimal`：

```go
if quote.RegularMarketPrice != nil {
    price := model.FromScaledDecimal(*quote.RegularMarketPrice)
    _ = price
}
```

一般 `facade.Quote` / `model.Quote` surface 已使用 `float64`，不需要自行換算 scale。

## Operator Checklist

1. 記錄 command、exit code、UTC timestamp 與 `--run-id`。
2. 執行 `yfin config --print-effective`，確認 config path 與 redaction。
3. 以單一 ticker / endpoint 重現。
4. 降低 `--qps` 排除 throttling；若是 `pull` universe / multi-ticker `quote`，再降低 `--concurrency`。
5. 對 parse failure 使用 `--preview-json` 或 `--preview-news`。
6. 修正後執行 unit tests，再在目標環境做 live provider verification。

## See Also

- [Configuration Reference](configuration.md)
- [Observability](observability.md)
- [CLI Usage](../cli/usage.md)
- [Scrape CLI](../scrape/cli.md)
- [Scrape Overview](../scrape/overview.md)
- [Operator Runbooks](../../monitoring/runbooks/)
