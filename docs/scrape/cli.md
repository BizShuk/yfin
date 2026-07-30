# CLI 使用指南 (CLI Usage Guide)

## 概觀 (Overview)

`yfin` CLI 提供 Yahoo Finance 網頁爬蟲的完整存取介面。本指南涵蓋所有爬蟲相關的 CLI 功能與實際範例。

`scrape` 為扁平子命令 (`yfin scrape`，無 nested subcommand)，每次執行需指定一種 mode。

## 命令結構 (Command Structure)

```bash
yfin [global-flags] scrape [flags]
```

## Scrape 命令

### 基本語法 (Basic Syntax)

```bash
yfin scrape [flags]
```

### Modes（互斥，一次只能指定一個）

`scrape` 命令提供三種執行模式，必須擇一指定：

| Mode | 用途 | 必要搭配旗標 |
| --- | --- | --- |
| `--check` | 連線測試（下載 response、不解析 DTO） | `--ticker` + `--endpoint` |
| `--preview-json` | extractor 乾跑（解析並輸出 JSON） | `--ticker` + `--endpoints` |
| `--preview-news` | news parser 乾跑 | `--ticker` |

> 模式旗標不可同時指定；若一個都沒給，`runScrape` 會回錯並退出。

### Core Flags

| Flag | Type | 說明 |
| --- | --- | --- |
| `--ticker` | string | 必填。股票代號（例如 `AAPL`） |
| `--endpoint` | string | 單一 endpoint，搭配 `--check` 使用。可選值：`profile`、`key-statistics`、`financials`、`balance-sheet`、`cash-flow`、`analysis`、`analyst-insights`、`news` |
| `--endpoints` | string | 多個 endpoint（逗號分隔），搭配 `--preview-json` 使用。可選值同上 |
| `--preview` | bool | 搭配 `--check` 顯示 response preview |
| `--preview-json` | bool | extractor 乾跑（mutually exclusive with other modes） |
| `--preview-news` | bool | news parser 乾跑（mutually exclusive） |
| `--check` | bool | 連線測試（mutually exclusive） |
| `--force` | bool | Deprecated compatibility flag；scrape 已是 explicit surface，目前無額外效果 |

> 註：CLI 不再提供任何 session-pool 旗標；HTTP client 由 `utils/httpx` 統一管理連線、限速與重試。

## 使用範例 (Usage Examples)

### 範例 1：基本 Key Statistics 連線測試

```bash
# 測試 Apple 的 key-statistics 頁面連線狀態
yfin scrape --ticker AAPL --endpoint key-statistics --check --preview

# 預期輸出：
# SCRAPE CHECK host=finance.yahoo.com url=https://finance.yahoo.com/quote/AAPL/key-statistics status=200 bytes=xxx gzip=true redirects=0 latency_p50≈xxxms
# CONTENT PREVIEW: <html>...</html>
```

**Why this example**：最常見的爬蟲情境——確認 Yahoo Finance 頁面可達、回應內容正常。

**When to use**：
- 部署前的 health check
- 排程故障排除時的第一步
- 確認 robots.txt / rate limit 沒有擋下請求

### 範例 2：多 endpoint JSON preview

```bash
# 一次預覽多個 endpoint 的 extractor 結果
yfin scrape --ticker MSFT --endpoints key-statistics,financials,analysis,profile --preview-json
```

**Why this example**：快速驗證多個 extractor 在同一個 ticker 上的解析狀況。

**When to use**：
- 開發新 extractor 時確認 parsing 結果
- 排查 schema drift
- 不需寫入下游 storage 的測試情境

### 範例 3：News parser preview

```bash
# 預覽 news 解析結果
yfin scrape --ticker TSLA --preview-news

# 預期輸出（摘要）：
# PREVIEW NEWS ticker=TSLA
# FETCH META: host=finance.yahoo.com status=200 bytes=xxx gzip=true ...
# TSLA news: found=15 deduped=2 returned=13 as_of=2024-01-03T16:45:00Z
# ARTICLES:
#  1) 2h ago   | Reuters         | Tesla Delivers Record Q4 Numbers
#  2) 1d ago   | Bloomberg       | Musk Announces New Gigafactory Plans
# ...
```

**Why this example**：驗證 news parser 對 dedup、relative time、related tickers 的處理。

**When to use**：
- 監控 news 解析行為
- 評估 sentiment analysis pipeline

`--force` 僅為 compatibility flag，不會停用 robots.txt enforcement；robots 行為仍由 `scrape.robots_policy` 決定。

## 進階使用模式 (Advanced Usage Patterns)

### 多 ticker 批次處理

`scrape` 命令本身不支援 `--universe-file`，需透過 shell 迴圈或 `soak` 命令批次處理：

```bash
#!/bin/bash
# batch-scrape.sh
TICKERS="AAPL MSFT GOOGL TSLA AMZN"

for ticker in $TICKERS; do
  echo "Processing $ticker..."
  yfin scrape --ticker "$ticker" --endpoint key-statistics --check
  sleep 1
done
```

### Rate limit 與 backoff 調整

`--qps`、`--retry-max`、`--timeout` 等為 HTTP global flags，需在 `yfin` 根命令指定：

```bash
yfin --qps 1.0 --timeout 45s scrape --ticker AAPL --endpoint key-statistics --check
```

> CLI 旗標優先於 Yahoo HTTP YAML config。Scrape-specific HTTP defaults 位於 Go config；Library caller 可傳入自訂 `*httpx.Config`。
>
> `--concurrency` 雖為 inherited persistent flag，但 `scrape` 沒有 multi-ticker worker pool，因此不 consume；它只控制 `pull` universe / multi-ticker `quote`。

## 錯誤排除與除錯 (Error Handling & Debugging)

### 除錯模式 (Debug Mode)

```bash
# 開啟 debug 等級 logging
yfin --log-level debug scrape --ticker AAPL --endpoint key-statistics --check
```

### Health Check Script

```bash
#!/bin/bash
# health-check.sh - 每日 endpoint 健康檢查

ENDPOINTS="profile key-statistics financials balance-sheet cash-flow analysis analyst-insights news"
TICKER="AAPL"

for endpoint in $ENDPOINTS; do
  echo "Checking $endpoint..."
  yfin scrape --ticker "$TICKER" --endpoint "$endpoint" --check
  if [ $? -eq 0 ]; then
    echo "✓ $endpoint healthy"
  else
    echo "✗ $endpoint failed"
  fi
done
```

## CLI 最佳實踐 (CLI Best Practices)

### 1. 配置管理 (Configuration Management)

```bash
# 使用環境對應的 config
yfin --config config/dev.yaml scrape --ticker AAPL --endpoint key-statistics --check
yfin --config config/prod.yaml scrape --ticker AAPL --endpoint key-statistics --check
```

### 2. 錯誤恢復 (Error Recovery)

```bash
# 透過 logging 與 run-id 追蹤問題
yfin --log-level debug --run-id "scrape-debug-$(date +%s)" \
  scrape --ticker AAPL --endpoint key-statistics --check 2>&1 | tee scrape.log
```

## 常見 CLI 模式 (Common CLI Patterns)

### 資料 pipeline 整合 (Data Pipeline Integration)

```bash
#!/bin/bash
# daily-scrape.sh - 每日資料蒐集腳本

TICKERS="AAPL MSFT GOOGL TSLA AMZN"
RUN_ID="daily-scrape-$(date +%Y%m%d)"

for ticker in $TICKERS; do
  echo "Processing $ticker..."

  yfin --config config/prod.yaml --run-id "$RUN_ID" \
    scrape --ticker "$ticker" \
    --endpoints key-statistics,financials,analysis,profile,news \
    --preview-json

  if [ $? -eq 0 ]; then
    echo "✓ $ticker completed"
  else
    echo "✗ $ticker failed"
  fi

  sleep 2
done
```

本指南涵蓋 `yfin scrape` 所有面向。詳細配置與疑難排解請參考其他章節。
