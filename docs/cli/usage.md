# Usage Guide

本指南提供 `yfin` 的完整使用範例 (This guide provides comprehensive examples of how to use `yfin` for fetching Yahoo Finance data).

## Overview

`yfin` 是擷取 Yahoo Finance 資料的命令列工具 (CLI tool)，支援：

- 每日 bars（調整後 / 原始）— 僅限 daily 區間 (Daily bars, adjusted/raw — daily-only scope)
- 即時 quote 快照 (Snapshot quotes)
- Fundamentals 季報（需付費訂閱）(Quarterly fundamentals — requires paid subscription)
- Yahoo Finance 網頁爬蟲（API 不可用時）(Web scraping fallback)
- TWSE（臺灣證券交易所）23 個 endpoint (TWSE 23 endpoints)
- Batch universe 擷取（skills/scripts parity）

## CLI Tree

`yfin` 採用 flat command 結構（無 nested groups）：

```tree
yfin
├── config              (cmd/admin/admin.go)
├── version             (cmd/admin/admin.go)
├── batch               (cmd/dispatch/batch.go)
├── fundamentals        (cmd/fundamentals/fundamentals_run.go)
├── comprehensive-stats (cmd/fundamentals/stats.go)
├── comprehensive-profile (cmd/fundamentals/profile.go)
├── pull                (cmd/market/pull.go)
├── quote               (cmd/market/quote.go)
├── scrape              (cmd/scrape/scrape_run.go)
└── twse                (cmd/twse/twse.go)
```

## Commands Overview

| Command | 用途 (Purpose) | 主要旗標 (Key Flags) |
| --- | --- | --- |
| `pull` | 擷取每日 bars（單 symbol 或 universe） | `--ticker` / `--universe-file` / `--start` / `--end` / `--adjusted` / `--out` / `--out-dir` |
| `quote` | 擷取即時 quote 快照 | `--tickers` (CSV) / `--out` / `--out-dir` |
| `fundamentals` | 擷取 fundamentals 季報（需付費） | `--ticker` / `--preview` |
| `comprehensive-stats` | 擷取完整 key statistics + 5 年歷史 | `--ticker` / `--preview` |
| `comprehensive-profile` | 擷取公司基本資料 + key executives | `--ticker` / `--preview` |
| `scrape` | Yahoo Finance 網頁爬蟲（3 種 mode） | `--check` / `--preview-json` / `--preview-news` / `--ticker` / `--endpoint` / `--endpoints` / `--force` |
| `twse` | 查詢任一 23 個 TWSE endpoint | `--endpoint` (必填) / `--date` / `--stock` / `--month` / `--timeout` / `--pretty` |
| `batch` | 批次擷取 universe 全部 commands | `--ticker` / `--max-workers` / `--force` |
| `config` | 印出 effective config | `--print-effective` / `--json` |
| `version` | 列印 CLI 版本與 build 細節 | — |

## Basic Usage

### Check Version

```bash
yfin version
```

### Get Help

```bash
# 通用說明 (general help)
yfin --help

# subcommand 說明 (command-specific help)
yfin pull --help
yfin quote --help
yfin fundamentals --help
yfin scrape --help
yfin twse --help
yfin config --help
```

## Daily Bars (`pull` command)

### Basic Bar Fetching

```bash
# 擷取單一 symbol 的每日 bars (fetch daily bars for a single symbol)
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# 指定 adjustment policy
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted raw --preview
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted split_dividend --preview
```

`--adjusted` 接受 `raw` 或 `split_dividend`，預設為 `split_dividend`。若 YAML 設定 `markets.default_adjustment_policy`，會作為省略 `--adjusted` 時的預設值。

### Multiple Symbols (Universe File)

```bash
# 建立 universe 檔 (create a universe file)
printf "AAPL\nMSFT\nGOOGL\nTSLA\n" > nasdaq100.txt

# 批次擷取 (fetch bars for multiple symbols)
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --preview
```

`--ticker` 與 `--universe-file` 互斥，僅能擇一。

### International Markets

```bash
# 德國市場 SAP (German market)
yfin pull --ticker SAP --market XETR --start 2024-01-01 --end 2024-12-31 --preview

# 日本市場 Toyota (Japanese market)
yfin pull --ticker TM --market XTKS --start 2024-01-01 --end 2024-12-31 --preview
```

### FX Conversion Preview

```bash
# 轉換為 EUR (convert to EUR)
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --fx-target EUR --preview

# 轉換為 JPY (convert to JPY)
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --fx-target JPY --preview
```

### Local Export

```bash
# 匯出 JSON (export to JSON)
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./data --preview

# 批次匯出多 symbol (export multiple symbols)
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./data --preview
```

### Local Export (Bars)

```bash
# 匯出 bars 至本地 JSON (export bars to local JSON)
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./out

# 批次匯出多 symbol (export multiple symbols)
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --out json --out-dir ./out
```

### Performance Tuning

```bash
# 提升 concurrency (increase concurrency)
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --concurrency 32 --preview

# 調整 QPS (queries per second)
yfin pull --universe-file nasdaq100.txt --start 2024-01-01 --end 2024-12-31 --qps 10 --preview
```

> **關於 `--sessions`**：此旗標保留為向後相容 (vestigial)，但 session rotation 已從 HTTP client 移除（見 `CLAUDE.md` 決策）。`--sessions` 對 connection reuse 不再生效，HTTP client 改採單一共享 `http.Client` + QPS rate limit + exponential backoff + circuit breaker。

## Snapshot Quotes (`quote` command)

### Single Quote

```bash
# 單 symbol quote
yfin quote --tickers AAPL --preview
```

### Multiple Quotes

```bash
# 多 symbol quote（CSV）
yfin quote --tickers AAPL,MSFT,GOOGL,TSLA --preview
```

### Export Quotes

```bash
# 匯出 quote 至 JSON
yfin quote --tickers AAPL,MSFT,GOOGL --out json --out-dir ./quotes --preview
```

### Local Export (Quotes)

```bash
# 匯出 quote 至本地 JSON (export quotes to local JSON)
yfin quote --tickers AAPL,MSFT --out json --out-dir ./out
```

## Fundamentals (`fundamentals` command)

**注意 (Note)**：fundamentals 需 Yahoo Finance 付費訂閱；未授權時 exit code 為 `2` (PaidFeatureRequired)。

### Basic Fundamentals

```bash
# 擷取 fundamentals 季報
yfin fundamentals --ticker AAPL --preview
```

### Comprehensive Variants

```bash
# 完整 key statistics（含 5 年歷史）
yfin comprehensive-stats --ticker AAPL

# 公司基本資料 + key executives
yfin comprehensive-profile --ticker MSFT --preview
```

### Error Handling

```bash
yfin fundamentals --ticker AAPL --preview
echo $?  # 401 / 未授權時為 2 (PaidFeatureRequired)
```

## Web Scraping (`scrape` command)

`scrape` 提供 3 種互斥 mode，必填其一：

```bash
# 連線測試 (connectivity check)
yfin scrape --check --ticker AAPL --endpoint profile --preview

# extractor 乾跑 (dry-run JSON extraction)
yfin scrape --preview-json --ticker AAPL --endpoints key-statistics,financials,analysis,profile

# news parser 乾跑 (dry-run news parser)
yfin scrape --preview-news --ticker AAPL
```

支援的 endpoint：`profile` / `key-statistics` / `financials` / `balance-sheet` / `cash-flow` / `analysis` / `analyst-insights` / `news`。

- `--check` mode：搭配 `--endpoint`（單一 endpoint）
- `--preview-json` mode：搭配 `--endpoints`（CSV，多 endpoint）
- `--force`：忽略 API 可用性檢查，強制走爬蟲

## TWSE (`twse` command)

23 個 TWSE endpoint 透過單一 `twse` 入口查詢：

```bash
# 加權指數 (MI_INDEX)
yfin twse --endpoint MI_INDEX --date 20221230

# 個股日成交（STOCK_DAY，需 --stock）
yfin twse --endpoint STOCK_DAY --date 20221230 --stock 2330

# 月報價（FMTQIK，需 --month）
yfin twse --endpoint FMTQIK --month 202212

# 週資料 (MI_WEEK)
yfin twse --endpoint MI_WEEK --date 20221230 --pretty
```

Endpoint 完整清單：`MI_INDEX` / `STOCK_DAY` / `BWIBBU_d` / `MI_INDEX_PLUS` / `MI_INDEX_ODD` / `MI_5MINS` / `TWTB4U` / `MI_MARGN` / `T86` / `MI_QFIIS` / `BFI82U` / `TWT38U` / `TWT43U` / `TWT44U` / `BFIAUU` / `BFIAUU_STOCK` / `BFIMUU` / `BFIAUU_YEAR` / `FMTQIK` / `STOCK_DAY_AVG` / `FMSRFK` / `BFIAMU` / `MI_WEEK`。

部分 endpoint 需要 `--stock`（個股代號）或 `--month`（`YYYYMM`）；未提供時 `twse` 會回 `unknown endpoint` / `missing --stock` / `missing --month` 錯誤。

## Batch (`batch` command)

批次擷取 universe 中所有 symbols 的所有 registered commands：

```bash
# 從 ticker_list.csv 讀取 universe
yfin batch

# 指定單一 ticker
yfin batch --ticker AAPL

# 強制重抓（忽略快取）
yfin batch --ticker AAPL --force

# 調整 worker pool 大小
yfin batch --max-workers 32
```

## Configuration Management

### View Effective Configuration

```bash
# 印出 effective config (key=value 格式)
yfin config --print-effective

# 以 JSON 格式印出
yfin config --print-effective --json
```

### Use Custom Configuration

```bash
# 指定 config 檔
yfin --config ./my-config.yaml pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Persistent Flags (every subcommand)

下表為 root 級 (persistent) 旗標，每個 subcommand 都可使用：

| 旗標 | 用途 |
| --- | --- |
| `--config` | YAML 設定檔路徑 |
| `--log-level` | Log level（`info` / `debug` / `warn` / `error`） |
| `--run-id` | Run ID（追蹤用；空白則自動產生） |
| `--concurrency` | Worker pool 大小（覆寫 YAML 預設） |
| `--qps` | Per-host QPS（覆寫 YAML 預設） |
| `--retry-max` | HTTP retry 嘗試次數 |
| `--timeout` | HTTP timeout（如 `6s`） |
| `--observability-disable-tracing` | 關閉 OpenTelemetry tracing |
| `--observability-disable-metrics` | 關閉 Prometheus metrics |

## Advanced Usage

### Observability Control

```bash
# 關閉 tracing
yfin --observability-disable-tracing pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# 關閉 metrics
yfin --observability-disable-metrics pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# 同時關閉
yfin --observability-disable-tracing --observability-disable-metrics pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Logging Control

```bash
# Debug logging
yfin --log-level debug pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# Warning level
yfin --log-level warn pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Custom Run ID

```bash
# 指定 run ID 用於追蹤
yfin --run-id my-daily-job pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### HTTP Configuration

```bash
# 自訂 timeout
yfin --timeout 30s pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# 自訂 retry 次數
yfin --retry-max 5 pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

## Output Examples

### Bar Preview Output

`cmd/market/pull.go` 的 `printBarsPreview` 輸出格式：

```
RUN <run-id>  (env=<env>, topic_prefix=<prefix>)
SYMBOL <symbol> (MIC=<mic>, CCY=<ccy>)  range=<start>..<end>  bars=<n>  adjusted=<policy>
first=<ts>  last=<ts>  last_close=<price> <ccy>
```

實例 (Example)：

```
RUN yfin_1704067200
SYMBOL AAPL (MIC=XNAS, CCY=USD)  range=2024-01-01..2024-12-31  bars=252  adjusted=split_dividend
first=2024-01-01T00:00:00Z  last=2024-12-31T00:00:00Z  last_close=192.5300 USD
```

### Quote Preview Output

`cmd/market/quote.go` 的 `printQuotePreview` 輸出格式：

```
SYMBOL AAPL quote  price=192.5300 USD  high=195.0000  low=190.0000  venue=XNAS
```

欄位缺失時顯示 `N/A`。

### Fundamentals Preview Output

`cmd/fundamentals/fundamentals_run.go` 的 `printFundamentalsPreview` 輸出前 5 行：

```
SYMBOL AAPL fundamentals  lines=45  source=yahoo-finance
  market_cap: 3000000000000.00 USD
  revenue: 383285000000.00 USD
  net_income: 99803000000.00 USD
  eps: 6.13 USD
  pe_ratio: 31.40
```

### Scrape Preview Output

`cmd/scrape/scrape_run.go` 的 `runScrapeCheck` 輸出：

```
SCRAPE CHECK host=finance.yahoo.com url=https://finance.yahoo.com/quote/AAPL/profile status=200 bytes=12345 gzip=true redirects=0 latency_p50≈1234ms
CONTENT PREVIEW: <html ...>
```

### Bus Preview Output

## Error Handling

### Exit Codes

定義於 `cmd/exitcodes.go`：

| Code | 名稱 | 說明 |
| --- | --- | --- |
| `0` | `ExitSuccess` | 成功 |
| `1` | `ExitGeneral` | 一般錯誤（網路、執行失敗等） |
| `2` | `ExitPaidFeature` | 付費功能未授權（fundamentals） |
| `3` | `ExitConfigError` | 組態錯誤（CLI flag 驗證失敗、YAML 載入失敗） |

### Common Error Scenarios

```bash
# 日期格式錯誤 (invalid date format)
yfin pull --ticker AAPL --start 2024-13-01 --end 2024-12-31 --preview
# ERROR: Invalid date format: parsing time "2024-13-01": month out of range

# 必填旗標缺少 (missing required flags)
yfin pull --ticker AAPL --start 2024-01-01 --preview
# ERROR: --start and --end are required

# adjustment policy 非法 (invalid adjustment policy)
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --adjusted invalid --preview
# ERROR: --adjusted must be 'raw' or 'split_dividend'

# 付費訂閱需求 (paid subscription required)
yfin fundamentals --ticker AAPL --preview
# ERROR: This endpoint requires Yahoo Finance paid subscription (exit 2)

# scrape mode 必填其一
yfin scrape --ticker AAPL
# ERROR: either --check, --preview-json, or --preview-news flag is required
```

## Best Practices

### Performance

1. **適當的 concurrency**：先使用預設值，universe 較大時再調升
2. **批次操作**：使用 universe 檔處理多 symbol
3. **監控 QPS**：避免超過 Yahoo Finance rate limit
4. **不要使用 `--sessions`**：session rotation 已移除，無效果；改用 `--qps` 與 `--concurrency`

### Data Quality

1. **務必先 preview**：正式發布前使用 `--preview`
2. **驗證日期區間**：確認 start/end 日期合理
3. **理解 adjustment policy**：`raw` vs `split_dividend` 影響 OHLCV 數值
4. **驗證 ticker 拼字**：確保 symbol 正確

### Production Usage

1. **使用 config 檔**：避免依賴 CLI flag 部署
2. **監控 observability**：啟用 tracing + metrics
3. **處理錯誤**：檢查 exit code 與 stderr
4. **環境分離**：使用 dev / staging / prod config

### Security

1. **保護設定檔**：機敏 config 檔應限制存取
2. **使用環境變數**：透過 `APP_*` 注入機敏設定
3. **定期更新**：保持 binary 為最新以修補漏洞

## Integration Examples

### Shell Scripts

```bash
#!/bin/bash
# 每日資料擷取腳本 (daily data fetch script)

DATE=$(date -d "yesterday" +%Y-%m-%d)
OUTPUT_DIR="./data/$(date +%Y/%m)"

yfin pull \
  --universe-file ./symbols/nasdaq100.txt \
  --start "$DATE" \
  --end "$DATE" \
  --out json \
  --out-dir "$OUTPUT_DIR" \
  --concurrency 16 \
  --log-level info

if [ $? -eq 0 ]; then
  echo "Data fetch completed successfully"
else
  echo "Data fetch failed"
  exit 1
fi
```

### Cron Jobs

```bash
# Crontab entry: 每日 06:00 擷取昨日資料
0 6 * * * /usr/local/bin/yfin pull --universe-file /path/to/symbols.txt --start $(date -d "yesterday" +\%Y-\%m-\%d) --end $(date -d "yesterday" +\%Y-\%m-\%d) --out json --out-dir /data/$(date +\%Y/\%m) --config /etc/yfin/config.yaml
```

### Docker Usage

```bash
# 於自建 Docker image 中執行 (run in custom Docker container)
docker run --rm \
  -v $(pwd)/config:/config \
  -v $(pwd)/data:/data \
  yfin:latest \
  yfin pull \
  --config /config/prod.yaml \
  --ticker AAPL \
  --start 2024-01-01 \
  --end 2024-12-31 \
  --out json \
  --out-dir /data
```

## Troubleshooting

### Common Issues

1. **網路逾時 (network timeouts)**：調升 `--timeout`，或檢查網路狀態
2. **Rate limiting**：降低 `--qps`，或減少 `--concurrency`
3. **記憶體問題**：大型 universe 時降低 `--concurrency`
4. **組態錯誤**：使用 `yfin config --print-effective` 除錯

### Debug Mode

```bash
# 啟用 debug logging
yfin --log-level debug pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview

# 關閉 observability 並啟用 debug
yfin --observability-disable-tracing --observability-disable-metrics --log-level debug pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Getting Help

```bash
# 檢查版本與 build 資訊
yfin version

# 詳細 subcommand 說明
yfin --help
yfin pull --help

# 驗證 config
yfin config --print-effective
```

## Next Steps

- [Installation Guide](../getting-started/install.md) - 如何安裝 `yfin`
- [Configuration](https://github.com/bizshuk/yfin/tree/main/configs) - 環境組態範例與選項