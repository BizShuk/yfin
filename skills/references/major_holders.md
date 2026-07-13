# major_holders

**Type:** `DataFrame` — 主要持股者資料

列出股票的主要持有人及其持股資訊。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Holder` | `str` | 持有人名稱 |
| `Shares` | `int` | 持股數量（股數） |
| `Date Reported` | `datetime` | 資料報告日期 |
| `% Held` | `float` | 持股比例（%） |
| `Value` | `float` | 持股市值 |

## Field Details

### Holder（持有人）
- 主要機構或大股東名稱
- 通常包括：創辦人、家族、子公司、或其他法人

### Shares（持股數量）
- 持有的股票數量（股）
- 數值可能為百萬股（M）或千股（K）等簡寫

### % Held（持股比例）
- 該持有人持有股份佔總發行股數的百分比
- 對於上市公司，通常以百分比表示

### Value（持股市值）
- 持股數量 × 當時股價
- 以股票交易的貨幣計算

## 持股者類型

| 類型 | 說明 |
|------|------|
| 內部人 | 董監事、高階主管及其關係人 |
| 法人 | 子公司、關係企業 |
| 員工 | 員工持股計劃 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
holders = ticker.major_holders
print(holders)
```

## 解讀要點

- 內部人持股過高可能影響股票流動性
- 內部人買入通常被視為正面信號
- 持股集中度過高可能造成大額交易的價格影響

## Notes

- 資料來源為 Yahoo Finance，可能與 SEC 文件略有差異
- 持股資料有報告延遲，並非即時資料
- 部分公司可能沒有主要持股者資料
