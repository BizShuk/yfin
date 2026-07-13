# calendar

**Type:** `DataFrame` — 行事曆事件資料

提供與股票相關的重要行事曆事件，如財報公布日期、股息支付日等。

## DataFrame Structure

欄位結構依事件類型而異，通常包含：

| Column | Type | Description |
|--------|------|-------------|
| `Earnings Date` | `datetime` | 財報公布日期 |
| `Earnings Jump` | `float` | 財報後的價格變動（%） |
| `Earnings Start` | `datetime` | 財報開始時間 |
| 其他事件欄位 | `varies` | 依可用資料 |

## Field Details

### Earnings Date（財報公布日期）
- 公司預定發布財報的日期
- 可能是預估日期，公司可能調整

### Earnings Jump（價格變動）
- 財報公布後的價格變動百分比
- 正值表示上涨，負值表示下跌
- 僅有歷史記錄有此數據

### Earnings Start（開始時間）
- 財報電話會議的開始時間
- 通常為美國東部時間

## 常見行事曆事件

| 事件類型 | 說明 |
|----------|------|
| Earnings | 財報公布 |
| ExDate | 除息日 |
| Dividend PayDate | 股息支付日 |
| Annual Meeting | 股東大會 |
| Conference Call | 財報電話會議 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
cal = ticker.calendar
print(cal)
```

## 重要提醒

| 事件 | 投資者需要注意 |
|------|---------------|
| 除息日 | 過此日期買入無法獲得股息 |
| 財報公布 | 可能導致大幅波動 |
| 電話會議 | 可獲取管理層對未來的看法 |

## Notes

- 並非所有股票都有完整的行事曆資料
- 未來事件的日期可能有所調整
- Yahoo Finance 的行事曆資料可能不完整
