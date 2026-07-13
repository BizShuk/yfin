# insider_transactions

**Type:** `DataFrame` — 內部人交易資料

記錄公司內部人（董事、監察人、高階主管）的股票交易歷史。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Name` | `str` | 內部人姓名 |
| `Transaction` | `str` | 交易類型（如 "Buy", "Sell"） |
| `Shares` | `int` | 交易股數（正數為買入，負數為賣出） |
| `Date` | `datetime` | 交易日期 |
| `Value` | `float` | 交易總值（股數 × 價格） |
| `Shares Total` | `int` | 該內部人目前持股總數 |

## Field Details

### Name（姓名）
- 內部人的完整姓名
- 通常是公司的董事、高階主管或大股東

### Transaction（交易類型）

| 類型 | 說明 |
|------|------|
| `Buy` | 買入股票 |
| `Sell` | 賣出股票 |
| `Option Exercise` | 行使股票期權 |
| `Gift` | 贈與股票 |
| `Rule 10b5-1` | 依據 10b5-1 計劃進行的交易 |

### Shares（股數）
- 正數：買入
- 負數：賣出
- 數量可能因股票分割而調整

### Value（交易值）
- `Shares × 交易價格`
- 以美元計算
- 反映交易的實際金額

### Shares Total（持股總數）
- 該內部人在交易後的總持股數
- 不包含尚未行使的期權

## 交易解讀

| 訊號 | 說明 |
|------|------|
| 內部人大量買入 | 對公司前景有信心 |
| 內部人大量賣出 | 可能存在風險或流動性需求 |
| 期權行使 | 通常為中性，需看後續是否賣出 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
insider = ticker.insider_transactions
# 找出近期的買入交易
buys = insider[insider['Transaction'] == 'Buy']
print(buys.tail(10))
```

## 法規要求

- 美國上市公司內部人需在交易後 2 個營業日內向 SEC 申報
- 10b5-1 計劃允許內部人在預先設定的條件下自動交易
- 內部人不得在掌握重大非公開資訊時交易

## Notes

- Yahoo Finance 的資料可能不完全或延遲
- 某些大額交易可能需要分多筆申報
- 員工股票購買計劃（ESPP）的交易可能單獨列出
