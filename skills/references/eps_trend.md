# eps_trend

**Type:** `DataFrame` — EPS 趨勢預測

提供分析師對未來 EPS 的趨勢預測，包括不同時間範圍的調整。

## DataFrame Structure

| Column | Type    | Description              |
| ------ | ------- | ------------------------ |
| `0`    | `float` | 當季/當年 EPS 共識預估   |
| `+7D`  | `float` | 7 天前的預估（變化趨勢） |
| `+30D` | `float` | 30 天前的預估            |
| `-30D` | `float` | 30 天後的預估（反向）    |
| `-7D`  | `float` | 7 天後的預估（反向）     |

## Field Details

### 0（當前期預估）

- 目前的華爾街共識 EPS 預估
- 代表當下分析師對該季度的最佳估計

### +7D / +30D（近期調整）

- 7 天前 / 30 天前的預估
- 用於觀察分析師共識的變化方向
- 若 +30D > 0，表示預估向上調整

### -30D / -7D（未來預估）

- 30 天後 / 7 天後的分析師預估
- 這些欄位的具體意義可能因資料來源而異

## 趨勢解讀

| 模式         | 說明             |
| ------------ | ---------------- |
| 預估持續上調 | 分析師越來越樂觀 |
| 預估持續下調 | 分析師越來越悲觀 |
| 預估穩定     | 市場共識一致     |

```python
# 計算 30 天 EPS 調整幅度
adjustment_30d = eps_trend['0'] - eps_trend['+30D']
adjustment_pct = (adjustment_30d / eps_trend['+30D']) * 100
print(f"30-Day EPS Revision: {adjustment_pct.iloc[0]:+.2f}%")
```

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
eps_t = ticker.eps_trend
print(eps_t)
```

## 與 EPS Revisions 的差異

| 特性     | `eps_trend`            | `eps_revisions`    |
| -------- | ---------------------- | ------------------ |
| 顯示方式 | 顯示多個時間點的預估值 | 顯示預估變化的統計 |
| 用途     | 觀察趨勢方向           | 量化調整幅度       |
| 解讀     | 直觀的走勢圖           | 具體的數字變化     |

## Notes

- 並非所有股票都有完整的趨勢資料
- 資料更新頻率為每日或每週
- 建議與 `eps_revisions` 配合分析
