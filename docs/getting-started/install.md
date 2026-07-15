# Installation Guide

本指南涵蓋 `yfin` 的三種安裝方式 (This guide covers three ways to install `yfin` on your system).

## Pre-built Binaries (Recommended)

### Download from GitHub Releases

1. 前往 [Releases 頁面](https://github.com/bizshuk/yfin/releases)
2. 下載對應平台的 binary：
   - `yfin_darwin_amd64.tar.gz` - macOS (Intel)
   - `yfin_darwin_arm64.tar.gz` - macOS (Apple Silicon)
   - `yfin_linux_amd64.tar.gz` - Linux (x86_64)
   - `yfin_linux_arm64.tar.gz` - Linux (ARM64)

3. 解壓並安裝：

```bash
# 解壓 (extract)
tar xzf yfin_darwin_arm64.tar.gz

# 移至 PATH 中的目錄 (move to a directory in your PATH)
sudo mv yfin /usr/local/bin/

# 驗證安裝 (verify)
yfin version
```

### Using curl (macOS/Linux)

```bash
# 設定變數 (set variables)
VERSION="v1.0.0"
OS="darwin"  # 或 "linux"
ARCH="arm64" # 或 "amd64"

# 下載並安裝 (download and install)
curl -L "https://github.com/bizshuk/yfin/releases/download/${VERSION}/yfin_${OS}_${ARCH}.tar.gz" | tar xz
sudo mv yfin /usr/local/bin/
yfin version
```

### Verify Installation

```bash
yfin version
```

預期輸出 (Expected output)：

```
yfin version v1.0.0
commit: abc123
build date: 2024-01-15
```

## Build from Source

### Prerequisites

- Go 1.26 或更新版本 (Go 1.26+)
- Git

### Build Steps

```bash
# 複製儲存庫 (clone the repository)
git clone https://github.com/bizshuk/yfin.git
cd yfin

# 編譯 binary (build the binary)
go build -o yfin ./cmd/yfin

# 安裝至系統 (install to system)
sudo mv yfin /usr/local/bin/

# 驗證安裝 (verify)
yfin version
```

### Build with Version Information

```bash
# 注入版本資訊 (build with version details)
go build -ldflags="-X main.version=v1.0.0 -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%d)" -o yfin ./cmd/yfin
```

## Go Module Installation

若要將 `yfin` 作為 Go 函式庫使用 (If you want to use `yfin` as a Go library)：

```bash
go get github.com/bizshuk/yfin@v1.0.0
```

於 Go 程式碼中 import (Then import in your Go code)：

```go
import "github.com/bizshuk/yfin"
```

> **Facade boundary 提醒 (Facade boundary reminder)**：對外 import 請走 `github.com/bizshuk/yfin/facade`，避免直接 import `internal/` 或 `svc/` 的型別。

## Docker

尚未發佈預編譯 Docker image (No pre-built Docker image is published at this time)。以下為自行 build 的範例 (Use the Dockerfile below to build a custom image)：

```dockerfile
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o yfin ./cmd/yfin

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/yfin .
CMD ["./yfin"]
```

Build 與執行 (Build and run)：

```bash
docker build -t yfin .
docker run --rm yfin yfin version
```

## Configuration

### Default Configuration

`yfin` 透過 `gosdk/config` 載入 YAML 設定；慣例 (convention) 路徑為：

- 應用設定根目錄：`~/.config/yfin/`
- 設定檔：`config/effective.yaml`（相對於執行目錄）
- 環境覆寫：`APP_*` 環境變數（由 `gosdk/config` 預設處理）

> `gosdk/config` 不接受自訂 config path；如需切換環境，請使用 `APP_*` 環境變數或切換工作目錄。

### Custom Configuration

```bash
# 指定 config 檔 (use custom config file)
yfin pull --config ./my-config.yaml --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

### Environment Variables (via gosdk)

`yfin` 不直接綁定 `YFINANCE_*` 環境變數；log / observability 設定透過 `gosdk` 的 `APP_*` 慣例覆寫：

```bash
export APP_LOG_LEVEL=debug
export APP_ENV=dev
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-12-31 --preview
```

CLI flag (`--log-level`, `--concurrency`, `--qps`, `--retry-max`, `--timeout`) 為執行期主要覆寫手段，env vars 僅在 `gosdk` 啟動時讀取 log level。

## Verification

### Test Installation

```bash
# 檢查版本 (check version)
yfin version

# 測試基本功能 (test basic functionality)
yfin pull --ticker AAPL --start 2024-01-01 --end 2024-01-02 --preview

# 印出 effective config (view effective configuration)
yfin config --print-effective
```

### Expected Output

```bash
$ yfin version
yfin version v1.0.0
commit: abc123
build date: 2024-01-15

$ yfin pull --ticker AAPL --start 2024-01-01 --end 2024-01-02 --preview
RUN yfin_1704067200  (env=dev)
SYMBOL AAPL (MIC=XNAS, CCY=USD)  range=2024-01-01..2024-01-02  bars=1  adjusted=split_dividend
first=2024-01-01T00:00:00Z  last=2024-01-02T00:00:00Z  last_close=192.5300 USD
```

## Troubleshooting

### Common Issues

#### Permission Denied

```bash
# 確認 binary 可執行 (make sure the binary is executable)
chmod +x yfin

# 檢查 PATH (check PATH)
echo $PATH
which yfin
```

#### Go Module Issues

```bash
# 清除 module cache (clean module cache)
go clean -modcache

# 更新相依套件 (update dependencies)
go mod tidy
go mod download
```

#### Configuration Issues

```bash
# 檢查組態 (check configuration)
yfin config --print-effective

# 從指定 config 載入 (validate configuration file)
yfin config --config ./config/effective.yaml --print-effective
```

### Getting Help

```bash
# 顯示說明 (show help)
yfin --help

# 顯示 subcommand 說明 (show command-specific help)
yfin pull --help
yfin quote --help
yfin fundamentals --help
yfin scrape --help
yfin twse --help
yfin config --help
yfin version --help
```

### Logs and Debugging

```bash
# 啟用 debug logging (enable debug logging)
yfin --log-level debug pull --ticker AAPL --start 2024-01-01 --end 2024-01-02 --preview

# 關閉 observability（測試用）(disable observability for testing)
yfin --observability-disable-tracing --observability-disable-metrics pull --ticker AAPL --start 2024-01-01 --end 2024-01-02 --preview
```

## Uninstallation

### Remove Binary

```bash
# 從系統移除 (remove from system)
sudo rm /usr/local/bin/yfin

# 或從使用者目錄移除 (or if installed in user directory)
rm ~/bin/yfin
```

### Remove Go Module

```bash
# 從 go.mod 移除 (remove from go.mod)
go mod edit -droprequire github.com/bizshuk/yfin

# 清理 (clean up)
go mod tidy
```

## Security

### Verify Binary Integrity

```bash
# 下載 checksums (download checksums)
curl -L "https://github.com/bizshuk/yfin/releases/download/v1.0.0/checksums.txt" -o checksums.txt

# 驗證下載的 binary (verify downloaded binary)
shasum -a 256 -c checksums.txt
```

### GPG Signatures

如有 GPG 簽章可供驗證 (If GPG signatures are available)：

```bash
# 匯入公鑰 (import public key)
gpg --keyserver keyserver.ubuntu.com --recv-keys <KEY_ID>

# 驗證簽章 (verify signature)
gpg --verify yfin_darwin_arm64.tar.gz.asc yfin_darwin_arm64.tar.gz
```

## Next Steps

安裝完成後，可參考 (After installation, see)：

- [Usage Guide](../cli/usage.md) - 學習如何使用 `yfin`
- [Configuration](https://github.com/bizshuk/yfin/tree/main/configs) - 環境組態範例