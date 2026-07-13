# earnings_history

**Type:** `DataFrame` — 盈餘歷史資料

記錄公司過往的盈餘表現歷史。

## DataFrame Structure

具體欄位依資料可用性而定，通常包含：

| Column | Type | Description |
|--------|------|-------------|
| `EPS Actual` | `float` | 實際 EPS |
| `EPS Estimate` | `float` | 預估 EPS |
| `EPS Surprise` | `float` | 與預估的差異 |
| `Earnings Date` | `datetime` | 公布日期 |

## Field Details

### EPS Actual（實際 EPS）
- 公司實際公布的每股收益
- 已根據股票分割調整

### EPS Estimate（預估 EPS）
- 公布前的華爾街共識預估
- 多個分析師的平均預測

### EPS Surprise（驚喜）
- `Actual - Estimate`
- 正值表示超預期（beat）
- 負值表示低於預期（miss）

## 與 earnings_dates 的差異

| 特性 | `earnings_history` | `earnings_dates` |
|------|-------------------|------------------|
| 時間範圍 | 僅歷史資料 | 含未來預期 |
| 用途 | 分析過去表現 | 預測未來事件 |
| 完整性 | 已公布的實際數據 | 可能有預估值 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
history = ticker.earnings_history
print(history.tail(10))
```

## 分析價值

```python
# 計算持續超預期的次數
beats = history[history['EPS Surprise'] > 0]
print(f"Quarterly Beats: {len(beats)}")

# 計算平均 EPS 驚喜
avg_surprise = history['EPS Surprise'].mean()
print(f"Average EPS Surprise: ${avg_surprise:.2f}")
```

## 季節性模式

部分公司有明顯的盈餘季節性：
- 科技股：Q4（聖誕旺季）通常最強
- 零售股：Q4（假日購物季）最重要
- 金融股：可能有獨特的公曆年度模式

## Notes

- 並非所有股票都有完整的盈餘歷史
- Yahoo Finance 的資料可能只涵蓋 2-4 年
- 建議同時參考 `earnings_dates` 取得未來事件
