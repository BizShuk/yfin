---
name: yf
description: 封裝 yfinance 程式庫以獲取股票數據、資訊和新聞 (Wrap yfinance library to fetch stock data, info and news).
---

# yf

封裝 `yfinance` 程式庫，做為純粹的資料擷取函式庫 (Data Fetching Library) 使用。

## 路徑 (Path)

```tree
skills/scripts/
├── yf.py              # 提供 yfinance 各指令的 API 接口 (30 指令)
├── all_ticker_yf.py   # 批次 fetch 主程序 (67 tickers × 30 指令)
├── compare_yfin_parity.py # Python oracle 與 Go artifacts 的 semantic comparator
└── config.py          # 共用設定 (COMMANDS, REFRESH_MAP)
```

## 前置需求 (Prerequisites)

1. Python 3.10+
2. `pip install yfinance pandas typer`
3. 確保虛擬環境 `~/.venv` 已安裝所需套件。

## 函式庫使用方式 (Library Usage)

`yf.py` 已轉為純 Python 模組，不提供命令行工具 (CLI)。其他 Python 腳本可以直接導入並調用 `fetch_data` 函式。

```python
import yf

# 獲取台積電的歷史價格數據 (預設為 30d 期間、1d 間隔)
history_json = yf.fetch_data("history", "2330.TW")

# 獲取其他自訂參數
income_json = yf.fetch_data("income", "2330.TW", freq="quarterly")
```

## 批次資料擷取 (Batch Fetch)

透過 `all_ticker_yf.py` 腳本，可以批次下載所有 Tickers 的各項數據，並存儲為 Raw JSON 格式：

```bash
# 下載 cmd/dispatch/ticker_list.csv 中所有個股的所有數據
~/.venv/bin/python skills/scripts/all_ticker_yf.py

# 僅下載單一股票
~/.venv/bin/python skills/scripts/all_ticker_yf.py --ticker 2330.TW

# 調整並行線程數 (預設為 10)
~/.venv/bin/python skills/scripts/all_ticker_yf.py --max-workers 5

# 無視快取，強制重新下載
~/.venv/bin/python skills/scripts/all_ticker_yf.py --force
```

- 輸出路徑：`~/.config/stock/data/raw/<command>/<ticker>.<YYYY-MM-DD>.json`
- 錯誤記錄：`~/.config/stock/data/raw/_failed/<ticker>.<command>.err`

對於特殊的 `history` 指令，預設會下載 `period="30d"`、`interval="1d"` 的天級價格數據。

## 維度分類 (Dimension Categories)

| 維度 | 說明 | 涵蓋資料 |
| --- | --- | --- |
| 價格 | 市場交易形成的價格數據 | OHLCV、履約價 |
| 基本面 | 公司財務體質與盈餘能力 | 財報、EPS、營收、盈餘日期 |
| 技術 | 基於價格/基本面計算的指標或趨勢 | EPS trend、price targets、analyst estimates |
| 資金 | 法人/機構的買賣與持股變化 | 機構持股、內部人交易、13F、除權息 |
| 消息 | 市場參與者的意見與資訊流 | 分析師建議、新聞、升降評 |
| 資訊 | 公司基本資料或低頻變動狀態 | info、sustainability、isin、metadata |

## 快取過期分級 (Refresh Frequency)

| 等級 | 頻率 | 說明 |
| --- | --- | --- |
| 日更新 | daily | 市場交易資料 (OHLCV、分析師建議、新聞) |
| 季度更新 | quarterly | 財報、EPS、營收預估、機構持股揭露 (13F)、除權息 |
| 月/週更新 | monthly | 內部人交易備案 (Form 4)、行事曆下次盈餘日期 |
| 年更新 | annually | ISIN 識別碼、選擇權履約價結構 (近乎永久) |

## Go client parity gate

`all_ticker_yf.py` 是 live Python oracle；Go production path 是 root-level `batch` command：

- Go 的 ordered manifest 與 Python `COMMANDS` 同為 30 個名稱；CI 使用 injected registry 驗證完整 lifecycle，不需外網。
- Go runtime 固定寫 `~/.config/yfin/data/raw/`；Python oracle 固定寫 `~/.config/stock/data/raw/`，兩者不混放。
- Go output 使用 atomic rename，cache 看最新有效 artifact；任何 `failed` command 令 CLI exit non-zero。
- Live parity 必須以 comparator 檢查 artifact existence、JSON validity、empty semantics 與 top-level type；command wiring 本身不代表 live Yahoo 驗證已通過。

```bash
# 對齊 all_ticker_yf.py 的預設行為
go run . batch

# 單股
go run . batch --ticker 2330.TW

# 強制重抓
go run . batch --ticker 2330.TW --force

# Python oracle + Go batch + semantic comparator
./scripts/verify-yf-parity.sh 2330.TW
```
