# yfin — Yahoo Finance Client for Go

`yfin` 是 Yahoo Finance 與 TWSE 市場資料的 Go SDK + flat CLI。它把 upstream HTTP、decode、validation、normalization 與 optional raw cache 收斂在同一個 client boundary，讓 downstream 專案不必各自維護 transport 或 persistence。

主要用途：

- 擷取 daily bars、snapshot quote、company info、market data 與 fundamentals。
- 明確使用 Yahoo web scrape 補足 API 沒有提供的 profile、statistics、analysis 與 news。
- 查詢 23 個 TWSE open-data endpoints。
- 以固定 command manifest 執行 30 個 yfinance-compatible foundation surfaces。
- 讓 downstream consumer 使用 yfin-owned read-through raw cache。

技術 architecture、package ownership 與 dependency rules 由 [CLAUDE.md](CLAUDE.md) 單一擁有。

## 能力邊界 (Capability Boundaries)

- `pull` 只支援 daily interval。
- Root CLI 不會從 API command 隱式切換到 scrape；scrape 必須明確執行。
- Standalone soak runner 的 `auto` strategy 只有 `fundamentals → financials` 具備等價 fallback；`quote` 與 `daily-bars` 沒有 scrape fallback。
- `--fx-target` 只驗證 source/target currency 是否相同；yfin 目前沒有 FX rate provider，不執行跨幣別 conversion。
- Local export 目前只支援 JSON。
- Plain SDK DTO 使用 `float64`；需要 `ScaledDecimal` 精度時，使用 `Fetch*Norm` normalized surface。
- `--sessions` 是 deprecated compatibility flag，沒有 runtime effect。
- Tracing flags 保留相容性，但目前 tracer 是 no-op；Prometheus metrics 與 structured logs 仍可使用。
- yfin 不擁有 message bus、publisher 或 protobuf emission pipeline。

## 安裝與首次啟動 (Install and First Run)

需求：Go `1.26.0` 或相容版本。

```bash
git clone https://github.com/bizshuk/yfin.git
cd yfin

# 建立 ~/.config/yfin/data 與 ~/.config/yfin/logs；不啟動程式
./run.sh

go build -o yfin .
./yfin --help
```

也可以安裝到 `GOBIN`：

```bash
go install github.com/bizshuk/yfin@latest
```

預設設定檔是 `config/effective.yaml`。完整 precedence 與 schema 請見 [Configuration Reference](docs/operations/configuration.md)。

## CLI 快速開始 (CLI Quick Start)

### Daily bars

```bash
go run . pull \
  --ticker AAPL \
  --start 2026-01-01 \
  --end 2026-07-01 \
  --adjusted split_dividend
```

寫入 JSON：

```bash
go run . pull \
  --ticker AAPL \
  --start 2026-01-01 \
  --end 2026-07-01 \
  --out json \
  --out-dir ./out
```

### Snapshot quote

```bash
go run . quote --tickers AAPL,MSFT
```

### Explicit scrape

```bash
# Connectivity check
go run . scrape --check --ticker AAPL --endpoint profile --preview

# Parse multiple scrape surfaces to JSON
go run . scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics,financials,analysis,profile

# Parse news
go run . scrape --preview-news --ticker AAPL
```

### Foundation batch

```bash
# One ticker across the shared 30-command manifest
go run . batch --ticker AAPL --max-workers 6
```

Batch artifacts 使用 yfin-owned atomic raw cache，預設位於 `~/.config/yfin/data/raw/`。

### TWSE

```bash
go run . twse --help
```

TWSE endpoint 所需的 `--date`、`--stock`、`--month` 依 endpoint metadata 驗證；請先用 help 查看當前 registry。

### Standalone soak runner

`soak` 不是 `yfin` subcommand：

```bash
go run ./cmd/soak \
  --universe-file tests/testdata/universe/soak.txt \
  --duration 10m \
  --concurrency 4 \
  --qps 1 \
  --failure-rate 0.05
```

詳細 strategy、failure injection 與 live acceptance boundary 請見 [Soak Testing Guide](docs/cli/soak-testing.md)。

## Go SDK 快速開始 (SDK Quick Start)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bizshuk/yfin/facade"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client := facade.NewClient()
	quote, err := client.FetchQuote(ctx, "AAPL", "example-run")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s %.2f %s\n", quote.Symbol, quote.Price, quote.Currency)
}
```

常用 surface：

| 需求 | Method |
| --- | --- |
| Plain daily bars | `FetchDailyBars` |
| Precision-preserving daily bars | `FetchDailyBarsNorm` |
| Plain quote | `FetchQuote` |
| Precision-preserving quote | `FetchQuoteNorm` |
| Company metadata | `FetchCompanyInfo` |
| Current market state | `FetchMarketData` |
| Quarterly fundamentals | `FetchFundamentalsQuarterly` |
| Scraped financials DTO | `ScrapeFinancialsData` |
| Scraped profile DTO | `ScrapeProfileData` |
| Scraped key-statistics DTO | `ScrapeKeyStatisticsData` |
| Shared foundation command | `FetchFoundation` |

Downstream cache ownership 與 constructors 請見 [API Reference](docs/api/reference.md) 及 [downstream cache ownership](docs/memory/2026-07-29-downstream-cache-ownership.md)。

## Data Flow

使用者只需選擇 API、explicit scrape、TWSE 或 cache surface；所有 transport 都由 facade client 統一建立：

```text
CLI / downstream SDK consumer
            |
          facade
       /     |      \
 Yahoo API  scrape   TWSE
       \     |      /
       model DTO / normalization
```

完整 package graph 與 allowed imports 請見 [CLAUDE.md](CLAUDE.md)。

## Configuration and Runtime Data

預設 runtime layout：

```text
~/.config/yfin/
├── data/    # raw cache 與 batch artifacts
└── logs/    # PM2 / redirected process logs
```

`run.sh` 只初始化 metadata directories，可重複執行，不會啟動 CLI 或 background process。yfin 的 structured logs 預設寫到 stderr；`logs/` 是 operator-owned redirection/PM2 target。

## Verification

Deterministic checks：

```bash
./run_tests.sh
go vet ./...
go build ./...
```

Live integration 需要可連線至外部 providers，必須明確 opt in：

```bash
./run_tests.sh --live-integration
```

## 文件導覽 (Documentation)

- [Onboarding Guide](docs/getting-started/onboarding.md)
- [CLI Commands](docs/cli/commands.md)
- [API Reference](docs/api/reference.md)
- [Scrape Overview](docs/scrape/overview.md)
- [Configuration Reference](docs/operations/configuration.md)
- [Observability Guide](docs/operations/observability.md)
- [Error Handling](docs/operations/error-handling.md)
- [Terminology](docs/terminology.md)
- [Pending Work](README.todo)
- [Changelog](docs/CHANGELOG.md)

## Contributing

修改前先讀 [CLAUDE.md](CLAUDE.md) 的 ownership 與 dependency contract。提交前至少執行：

```bash
go fmt ./...
go test ./...
go vet ./...
```

`go test ./...` 會執行 internal dependency boundary assertion。

## License and Security

本專案使用 [MIT License](LICENSE)。安全問題請依 [Security Policy](SECURITY.md) 私下回報。
