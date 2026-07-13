# history

**Type:** `DataFrame` — 歷史股價資料（日期為索引）

取得個股的歷史價格資料，包含開盤價、收盤價、最高價、最低價、成交量與除權息資料。

## DataFrame Structure

| Column         | Type    | Description                                 |
| -------------- | ------- | ------------------------------------------- |
| `Open`         | `float` | 開盤價，該交易日第一筆成交價格              |
| `High`         | `float` | 最高價，該交易日最高成交價格                |
| `Low`          | `float` | 最低價，該交易日最低成交價格                |
| `Close`        | `float` | 收盤價，該交易日最後一筆成交價格            |
| `Volume`       | `int`   | 成交量，該日總成交股數                      |
| `Dividends`    | `float` | 股息，該日發放的每股股息金額（無則為0）     |
| `Stock Splits` | `float` | 股票分割比例（如 2.0 表示 1 股分割為 2 股） |

## Field Details

### Open（開盤價）

- 該交易日第一筆交易的成交價格
- 可能在盤前/盤後交易時段有不同的開盤價

### High（最高價）

- 該交易日內所有成交價格的最高值
- 不區分正常交易或盤前/盤後

### Low（最低價）

- 該交易日內所有成交價格的最低值
- 不區分正常交易或盤前/盤後

### Close（收盤價）

- 該交易日最後一筆交易的成交價格
- 通常是最重要的參考價格

### Volume（成交量）

- 該交易日總成交股數
- 是衡量股票流動性的關鍵指標

### Dividends（股息）

- 該日發放的每股股息
- 大多數日期此值為 0，只在除息日有值
- 可用於計算股息殖利率

### Stock Splits（股票分割）

- 分割比例，如 2.0 表示 1 股變成 2 股
- 反向分割則為 0.5（2 股合併為 1 股）
- 發生分割時，歷史價格會自動調整

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
hist = ticker.history(period="1mo")
print(hist.tail())
```

## Parameters

| Parameter  | Type       | Default | Description                                                  |
| ---------- | ---------- | ------- | ------------------------------------------------------------ |
| `period`   | `str`      | `"1mo"` | 資料期間（1d, 5d, 1mo, 3mo, 6mo, 1y, 2y, 5y, 10y, ytd, max） |
| `interval` | `str`      | `"1d"`  | 資料間隔（1m, 2m, 5m, 15m, 30m, 60m, 1h, 1d, 1wk, 1mo）      |
| `start`    | `datetime` | `None`  | 開始日期                                                     |
| `end`      | `datetime` | `None`  | 結束日期                                                     |

## Notes

- 歷史資料預設會自動處理股票分割（split-adjusted）
- `Dividends` 欄位記錄的是「每股」股息，非總股息
- 短週期資料（1m, 2m）僅適用於近期的資料
