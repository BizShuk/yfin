# insider_purchases

**Type:** `DataFrame` — 內部人購買股票資料

專門記錄公司內部人買入股票的交易。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Name` | `str` | 內部人姓名與頭銜 |
| `Transaction` | `str` | 交易類型 |
| `Shares` | `int` | 買入股數 |
| `Date` | `datetime` | 交易日期 |
| `Value` | `float` | 交易總值 |

## Field Details

### Name（姓名）
- 內部人姓名
- 有時包含職位（如 "CEO - John Doe"）

### Transaction（交易類型）
- 通常為 "Buy" 或 "Purchase"
- 可能包含更詳細的類型如 "Rule 10b5-1 Purchase"

### Shares（股數）
- 買入的股數
- 數量為正數

### Date（日期）
- 交易執行日期
- 通常為成交日

### Value（交易值）
- 買入金額 = 股數 × 股價
- 以美元計算

## 與 insider_transactions 的差異

| 特性 | `insider_purchases` | `insider_transactions` |
|------|---------------------|------------------------|
| 範圍 | 僅買入 | 包含買入與賣出 |
| 用途 | 專門分析內部人買入 | 全面分析所有內部人交易 |
| 記錄數 | 較少 | 較多 |

## 為何關注內部人買入

```
內部人買入被視為對公司最有信心的訊號
因為內部人使用自己的資金買入，與單純獲得期權不同
```

## Example

```python
import yfinance as yf
ticker = yf.Ticker("TSLA")
purchases = ticker.insider_purchases
# 計算近一年內部人總買入金額
total_bought = purchases[purchases['Date'] > '2025-01-01']['Value'].sum()
print(f"Total insider purchases: ${total_bought:,.0f}")
```

## 重要提醒

| 考量 | 說明 |
|------|------|
| 時間延遲 | 資料可能有數週延遲 |
| 規模考量 | 小公司的大額買入較有意義 |
| 持續性 | 一次性買入不如持續買入重要 |

## Notes

- 此資料為 `insider_transactions` 的子集（僅過濾買入）
- 部分股票可能沒有此資料
- 建議與 `insider_transactions` 合併分析以獲得完整圖像
