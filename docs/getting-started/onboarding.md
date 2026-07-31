# yfin 入門指南 (Onboarding Guide)

本指南提供從 clone 到第一個 verified change 的最短路徑。Business scope 與 quick start 由 [README](../../README.md) 擁有；architecture、directory ownership 與 dependency direction 由 [CLAUDE](../../CLAUDE.md) 擁有。

## 1. 準備環境 (Prepare the Environment)

Go toolchain 版本以 `go.mod` 為準：

```bash
go version
go mod download
```

初始化 metadata directories：

```bash
./run.sh
```

這個 script 只建立 `${XDG_CONFIG_HOME:-$HOME/.config}/yfin/data` 與 `logs`，不會啟動 CLI 或 background process。Runtime YAML 預設讀取 `config/effective.yaml`；也可用 global `--config` 指定其他檔案。

## 2. 建置與 Deterministic Verification

```bash
go build -o yfin .
./yfin --help
go test ./...
go test -tags=integration ./... -run '^$'
go vet ./...
```

`-run '^$'` 只確認 integration-tagged packages 能編譯，不會執行 live provider requests。要在可連線環境執行 live integration：

```bash
go test -tags=integration ./...
```

也可使用統一 test runner：

```bash
./run_tests.sh
./run_tests.sh --race
./run_tests.sh --live-integration
```

Live verification 依賴 Yahoo Finance / TWSE 的可達性與當下 upstream behavior，應與 deterministic result 分開記錄。

## 3. 認識 Public Surface

常用入口有兩組：

- `facade.Client`：HTTP、scrape、TWSE 與 cache operations 的 public handler。
- `model.*`：plain DTO、normalized DTO、`ScaledDecimal` 與 normalization helpers。

`facade.Bar`、`facade.Quote` 等資料型別是 `model.*` 的 compatibility aliases。Plain surface 使用 `float64`；只有 `Fetch*Norm` methods 回傳 precision-preserving `ScaledDecimal`。

最小 SDK 範例：

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
	client := facade.NewClient()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	batch, err := client.FetchDailyBars(
		context.Background(),
		"AAPL",
		start,
		end,
		true,
		"onboarding",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fetched %d bars\n", len(batch.Bars))
}
```

更多 executable examples：

- [`facade/samples/`](../../facade/samples/)
- [`cmd/samples/`](../../cmd/samples/)

## 4. 使用 CLI

CLI 採 flat command tree；目前 command 與 flag 以 `--help` 為 runtime source of truth：

```bash
./yfin version
./yfin config --print-effective
./yfin pull --help
./yfin quote --help
./yfin scrape --help
./yfin twse --help
```

Daily bars：

```bash
./yfin pull \
  --ticker AAPL \
  --start 2024-01-01 \
  --end 2024-12-31
```

Local export 只支援 JSON，`--out` 與 `--out-dir` 必須成對：

```bash
./yfin pull \
  --ticker AAPL \
  --start 2024-01-01 \
  --end 2024-12-31 \
  --out json \
  --out-dir ./data
```

`--fx-target` 只驗證 fetched currency 已符合 target。yfin 沒有 FX rate provider，因此不會執行跨幣別 conversion。

Scrape 是 explicit surface，不是 API command 的 general fallback：

```bash
./yfin scrape \
  --preview-json \
  --ticker AAPL \
  --endpoints key-statistics,profile
```

## 5. 進行 Code Change

先在 [CLAUDE](../../CLAUDE.md) 找到 target package 的 owner 與 allowed imports。共同規則：

- `cmd/*` 的 runtime fetch 只走 `facade`，不得直接 import `svc/*`。
- `model/` 是 lowest layer，不 import 其他 internal packages。
- Provider-specific HTTP / parsing 放在對應 `svc/*`。
- 跨 package 的 shared CLI formatting 放在 `cmd/format`，不要 copy-paste。
- 一個 executable assertion 應寫成 test 或 script，不以 Markdown 代替。

新增或修正 behavior 時，先加入會失敗的 focused test，再做最小 implementation。完成後執行：

```bash
go fmt ./...
go test ./...
go test -tags=integration ./... -run '^$'
go vet ./...
go build ./...
```

Architecture import contract 由 deterministic test 驗證：

```bash
go test ./tests/unit -run TestInternalPackagesFollowDependencyContract
```

此 test 涵蓋 `cmd`、`facade`、`svc`、`config`、`model` 與 `utils` 的 internal import boundary。

## 6. Debug 與 Observability

- CLI 以 structured `slog` 寫到 stderr；固定 log file、rotation 與 retention 由 deployment owner 管理。
- `utils/obsv` 提供 active HTTP Prometheus metrics。
- OpenTelemetry tracer 目前是 no-op boundary，尚未接 exporter。
- `cmd/soak` 是 standalone binary，以 `go run ./cmd/soak` 執行，不是 `yfin soak` subcommand。

遇到 provider failure 時，先縮小到一個 symbol / endpoint，降低 `--qps`，並保留 exact error。Integration failure 若只發生在 live suite，需先區分 upstream availability 與 deterministic regression。

## 7. Canonical References

- [README](../../README.md)：business definition 與 quick start
- [CLAUDE](../../CLAUDE.md)：architecture、ownership 與 hard decisions
- [Terminology](../terminology.md)：domain 名詞的 single source of truth
- [CLI Commands](../cli/commands.md)：command / flag reference
- [Configuration](../operations/configuration.md)：YAML 與 runtime settings
- [Error Handling](../operations/error-handling.md)：exit codes 與 troubleshooting
- [Observability](../operations/observability.md)：logs、metrics 與 tracing boundary
