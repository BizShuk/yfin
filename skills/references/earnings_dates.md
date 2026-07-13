# earnings_dates

**Type:** `DataFrame` — 盈餘公布日期資料

記錄公司過往與預期的盈餘（財報）公布日期及 EPS 數據。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Earnings Date` | `datetime` | 盈餘公布日期時間 |
| `EPS Estimate` | `float` | 分析師預估 EPS |
| `EPS Actual` | `float` | 實際 EPS |
| `Difference` | `float` | 實際與預估差異（beat/miss） |
| `Surprise %` | `float` | 驚喜百分比 |

## Field Details

### Earnings Date（公布日期）
- 財報發布的日期和時間
- 時間可能是 UTC 或當地時間
- 公布後股價通常會有大幅波動

### EPS Estimate（預估 EPS）
- 分析師在公布前的共識預估
- 華爾街共識通常是多個分析師的平均

### EPS Actual（實際 EPS）
- 公司實際公布的稀釋每股收益
- 季度或年度數據

### Difference（差異）
- `Actual - Estimate`
- 正數：超預期（Beat）
- 負數：低於預期（Miss）

### Surprise %（驚喜百分比）
- `(Actual - Estimate) / |Estimate| × 100%`
- 超過 0% 表示 beat consensus

## 資料類型

| 類型 | 說明 |
|------|------|
| 已公布 | 有實際 EPS 的歷史記錄 |
| 預期 | 尚未公布的未來日期 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
earnings_dates = ticker.earnings_dates
# 查看近期的盈餘公布
upcoming = earnings_dates[earnings_dates['EPS Actual'].isna()]
print(upcoming)
```

## 盈餘關鍵指標

```python
# 歷史 beat rate
actual_earnings = earnings_dates.dropna(subset=['EPS Actual'])
beat_count = (actual_earnings['Difference'] > 0).sum()
total_earnings = len(actual_earnings)
beat_rate = beat_count / total_earnings
print(f"Beat Rate: {beat_rate:.2%}")

# 平均驚喜幅度
avg_surprise = actual_earnings['Surprise %'].mean()
print(f"Average Surprise: {avg_surprise:.2f}%")
```

## 股價反應模式

| EPS 結果 | 典型股價反應 |
|----------|--------------|
| Beat + 上調指引 | 大幅上漲 > +5% |
| Beat + 下調指引 | 小幅上漲或持平 |
| Miss | 大幅下跌 > -5% |
| Miss + 下調指引 | 暴跌 |

## Notes

- 未來的公布日期可能只有預估，尚未確認
- 不同財報季度有不同的重要性（Q4 通常最重要）
- 公布後 EPS Actual 欄位才會有值
