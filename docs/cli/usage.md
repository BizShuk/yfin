# CLI Usage Guide

`yfin` 使用 flat command tree。Exact flags / defaults 以 `--help` 與 [Commands Reference](commands.md) 為準；本頁提供 common workflows。

## Discover the CLI

```bash
yfin version
yfin --help
yfin pull --help
yfin quote --help
yfin fundamentals --help
yfin scrape --help
yfin twse --help
yfin batch --help
yfin config --help
```

以下範例統一把 global flags 放在 subcommand 前（parser 也接受放在後方）：

```bash
yfin \
  --config ./config/effective.yaml \
  --log-level debug \
  --run-id manual-check \
  quote \
  --tickers AAPL
```

## Daily Bars

Single symbol：

```bash
yfin pull \
  --ticker AAPL \
  --start 2026-01-01 \
  --end 2026-01-31
```

Adjustment policy：

```bash
yfin pull \
  --ticker AAPL \
  --start 2026-01-01 \
  --end 2026-01-31 \
  --adjusted raw
```

`raw` 與 `split_dividend` 是唯一有效值；Cobra default 是 `split_dividend`。

Universe file：

```bash
printf 'AAPL\nMSFT\nGOOGL\n' > symbols.txt

yfin \
  --concurrency 8 \
  --qps 2 \
  pull \
  --universe-file symbols.txt \
  --start 2026-01-01 \
  --end 2026-01-31
```

Fetch 可 parallel，但 output 依輸入順序處理。Effective worker count 仍受 YAML global / per-host limits 與 item count 限制。

### MIC Hint

```bash
yfin pull \
  --ticker SAP \
  --start 2026-01-01 \
  --end 2026-01-31 \
  --market XETR
```

`--market` 可接受 MIC 或 known Yahoo exchange code（例如 `NMS → XNAS`），只在 provider MIC inference 留空時補 output metadata；不覆寫已知 MIC、不選擇 endpoint，也不改寫 symbol。

### Currency Assertion

```bash
yfin pull \
  --ticker AAPL \
  --start 2026-01-01 \
  --end 2026-01-31 \
  --fx-target USD
```

這是 assertion，不是 conversion。同幣別通過；跨幣別回 error，因為 production CLI 沒有 FX rate provider。

### JSON Export

```bash
yfin pull \
  --ticker AAPL \
  --start 2026-01-01 \
  --end 2026-01-31 \
  --out json \
  --out-dir ./data
```

只支援 JSON。`--out` 與 `--out-dir` 必須成對出現。

## Quotes

```bash
yfin quote --tickers AAPL
yfin --concurrency 4 quote --tickers AAPL,MSFT,GOOGL
yfin quote --tickers AAPL,MSFT --out json --out-dir ./quotes
```

Quote export 同樣只支援 JSON，且 output flags 必須成對。

## Fundamentals

Quarterly fundamentals：

```bash
yfin fundamentals --ticker AAPL --preview
```

此 Yahoo surface 可能需要 paid subscription；未授權時使用 dedicated exit code。

Explicit web surfaces：

```bash
yfin comprehensive-stats --ticker AAPL --preview
yfin comprehensive-profile --ticker AAPL --preview
```

Annual SDK methods 不是 CLI subcommands；Go consumer 使用 facade API。

## Scrape

一次必須選一個 mode：

```bash
# Connectivity only
yfin scrape --check --ticker AAPL --endpoint profile --preview

# Parse selected DTOs
yfin scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics,financials,analysis,profile

# News parser
yfin scrape --preview-news --ticker AAPL
```

Scrape 是 explicit surface。`--force` 是 compatibility flag，不會關閉 robots enforcement。完整說明見 [Scrape CLI](../scrape/cli.md)。

## TWSE

```bash
yfin twse --endpoint MI_INDEX --date 20221230
yfin twse --endpoint STOCK_DAY --date 20221230 --stock 2330
yfin twse --endpoint FMTQIK --month 202212
```

Endpoint-specific required flags 由 registry / dispatcher validation 擁有。先用 `yfin twse --help` 查看 current surface。

## Foundation Batch

```bash
yfin batch
yfin batch --ticker AAPL
yfin batch --ticker AAPL --force
yfin batch --max-workers 16
```

預設 universe embedded 在 binary；artifacts 寫入 `~/.config/yfin/data/raw/`。任一 failed command 使 CLI 回 non-zero，已成功的 atomic artifacts 保留。

Live semantic parity gate：

```bash
./scripts/verify-yf-parity.sh AAPL
```

它依賴 external provider / oracle，需與 deterministic tests 分開記錄。

## Configuration

```bash
yfin config --print-effective
yfin config --print-effective --json
yfin --config ./config/effective.yaml config --print-effective
```

Output 會 redact secret references。Runtime directories 可先建立：

```bash
./run.sh
```

## HTTP and Observability Controls

```bash
yfin \
  --qps 1 \
  --retry-max 3 \
  --timeout 20s \
  --log-level debug \
  pull \
  --ticker AAPL \
  --start 2026-01-01 \
  --end 2026-01-31
```

- `--sessions` 保留相容但無效果。
- `--observability-disable-metrics` 關閉 Prometheus server。
- `--observability-disable-tracing` 是 compatibility flag；tracer 目前本來就是 no-op。
- `--concurrency` 只影響 multi-symbol `pull` / `quote`。

## Exit Codes

| Code | Meaning |
| --- | --- |
| `0` | Success |
| `1` | General runtime / provider error |
| `2` | Paid feature required |
| `3` | Config / flag validation error |

詳細分類見 [Error Handling](../operations/error-handling.md)。

## Standalone Soak

Soak 不是 `yfin` subcommand：

```bash
go run ./cmd/soak --help
```

見 [Soak Testing](soak-testing.md)。
