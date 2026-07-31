# Installation Guide

目前 repository-owned installation paths 是 source build 與 `go install`。本 repo 沒有 Dockerfile、Homebrew formula 或 release-build workflow，因此本頁不宣稱存在預建 artifacts。

## Prerequisites

Go version 以 `go.mod` 為準：

```bash
go version
```

## Build from Source

```bash
git clone https://github.com/BizShuk/yfin.git
cd yfin
go mod download
go build -o yfin .
./yfin version
./yfin --help
```

需要放入 user PATH 時：

```bash
install -m 0755 ./yfin "${HOME}/bin/yfin"
```

確保 `${HOME}/bin` 已在 `PATH`。是否安裝到 system-wide directory 由 operator 自行決定。

## Go Install

指定已發布 tag：

```bash
go install github.com/bizshuk/yfin@<version>
```

Binary 會寫入 `GOBIN`；未設定時通常是 `$(go env GOPATH)/bin`。先確認：

```bash
go env GOBIN GOPATH
```

不要把文件中的 placeholder `<version>` 原樣執行；依 repository tags / release policy 選明確版本。

## SDK Dependency

在 Go application 中：

```bash
go get github.com/bizshuk/yfin@<version>
```

Operations 從 `facade` import，DTO 可從 `model` import：

```go
import (
	"github.com/bizshuk/yfin/facade"
	"github.com/bizshuk/yfin/model"
)
```

不要直接依賴 `svc/*` transport internals。

## Runtime Metadata Directories

Source checkout 提供 metadata-only setup：

```bash
./run.sh
```

它建立：

```text
${XDG_CONFIG_HOME:-$HOME/.config}/yfin/data
${XDG_CONFIG_HOME:-$HOME/.config}/yfin/logs
```

Script 不會啟動 CLI、PM2 或 background service。yfin logs 寫到 stderr；`logs/` 是 deployment metadata location，不代表 CLI 會自行寫 log file。

## Configuration

CLI 預設讀取 current working directory 下的 `config/effective.yaml`。Installed binary 應明確指定 deployment config：

```bash
yfin \
  --config /absolute/path/to/effective.yaml \
  config \
  --print-effective
```

範例統一把 global flags 放在 subcommand 前；Cobra parser 也接受放在 subcommand 後。YAML schema 與 precedence 見 [Configuration](../operations/configuration.md)。

## Build Metadata

Default source build 會顯示 `dev / unknown`。需要注入 metadata：

```bash
go build \
  -ldflags "\
-X github.com/bizshuk/yfin/cmd.Version=<version> \
-X github.com/bizshuk/yfin/cmd.Commit=$(git rev-parse --short HEAD) \
-X github.com/bizshuk/yfin/cmd.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o yfin .
```

## Verification

Offline / deterministic checks：

```bash
./yfin version
./yfin --help
go test ./...
go test -tags=integration ./... -run '^$'
```

Live smoke（需要 provider network）：

```bash
./yfin \
  --config ./config/effective.yaml \
  pull \
  --ticker AAPL \
  --start 2026-01-01 \
  --end 2026-01-05
```

Live fetch failure 不等於 installation failure；先確認 `version`、`--help`、config load 與 deterministic tests。

## Troubleshooting

- `command not found`：檢查 `GOBIN` / `GOPATH/bin` / `PATH`。
- config not found：用 global `--config` 傳入明確路徑。
- provider timeout / 429：依 [Error Handling](../operations/error-handling.md) 降低 QPS 或調整 timeout。
- module download failure：確認 network、Go proxy / private-module settings 與 requested tag。

## See Also

- [Onboarding](onboarding.md)
- [CLI Usage](../cli/usage.md)
- [Versioning](../integrations/versioning.md)
- [Security Policy](../../SECURITY.md)
