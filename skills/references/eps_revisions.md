# eps_revisions

**Type:** `DataFrame` — EPS 修正資料

記錄分析師對 EPS 預估的修正情況。

## DataFrame Structure

具體欄位依資料可用性而定，通常包含：

| Column | Type | Description |
|--------|------|-------------|
| `Period` | `str` | 季度/年度標識 |
| `eps_revisions` | `dict` | 各項修正統計 |

## Field Details

### 修正類型

| 術語 | 說明 |
|------|------|
| Up Revisions | 向上修正的數量 |
| Down Revisions | 向下修正的數量 |
| Last Week | 一週前的預估 |
| Last Month | 一月前的預估 |
| Next Fiscal Year | 下一財年預估 |

## 與 eps_trend 的差異

| 特性 | `eps_trend` | `eps_revisions` |
|------|-------------|------------------|
| 呈現方式 | 預估值隨時間變化 | 修正次數統計 |
| 核心指標 | 預估數值 | 修正方向與數量 |
| 用途 | 了解共識如何移動 | 評估分析師信心變化 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
revisions = ticker.eps_revisions
print(revisions)
```

## 投資解讀

```python
# 計算修正方向比率
up_ratio = revisions['Up Revisions'] / (revisions['Up Revisions'] + revisions['Down Revisions'])
print(f"Revision Up Ratio: {up_ratio:.2%}")

# 趨勢分析
# 連續上調 = 強勁基本面
# 連續下調 = 基本面惡化警訊
```

## 為何關注 EPS 修正

1. **領先指標**：分析師修正通常領先實際盈餘
2. **趨勢確認**：持續上調強化買進理由
3. **逆勢訊號**：過度悲觀後的上調可能是買點

## Notes

- 並非所有股票都有完整的修正資料
- Yahoo Finance 可能只顯示彙總後的數據
- 建議與 `eps_trend` 配合以獲得完整圖像
