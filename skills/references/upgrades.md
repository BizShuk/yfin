# upgrades

**Type:** `DataFrame` — 分析師升評/降評資料

記錄分析師對股票評級的變動（升評或降評）。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Grade` | `str` | 最新評級 |
| `From` | `str` | 原評級 |
| `To` | `str` | 新評級 |
| `Action` | `str` | 動作（"upgrade" 或 "downgrade"） |
| `Date` | `datetime` | 公告日期 |
| `Firm` | `str` | 分析師所屬機構 |

## Field Details

### Grade（評級）
- 改變後的新評級
- 格式依機構而異

### From（原評級）
- 變動前的評級
- 可為空白（首次覆蓋）

### To（新評級）
- 變動後的評級
- 目標價格或評級改變

### Action（動作類型）

| 值 | 說明 |
|----|------|
| `upgrade` | 升評，評級變好 |
| `downgrade` | 降評，評級變差 |
| `initiated` | 首次覆蓋 |
| `reiterated` | 重申，維持不變 |

### Date（日期）
- 評級變動的公告日期
- 通常為交易日的盤前或盤後

### Firm（機構名稱）
- 發布研究報告的券商/投行
- 例如："Morgan Stanley", "Goldman Sachs"

## 為何關注升評/降評

| 事件 | 市場解讀 |
|------|----------|
| 升評 | 正面訊號，分析師認為將上漲 |
| 降評 | 負面訊號，分析師認為將下跌 |
| 首次覆蓋 | 新關注，代表有新的研究興趣 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
upgrades = ticker.upgrades
# 查看近期的評級變動
print(upgrades[upgrades['Date'] > '2025-01-01'])
```

## 影響市場的原因

1. **名聲效應**：大型券商的評級影響機構投資決策
2. **目標價**：隱含 10-20% 的潛在漲幅
3. **信心變化**：反映分析師對公司前景的看法

## Notes

- 並非所有股票都有評級變動記錄
- Yahoo Finance 通常保留 12-18 個月的歷史
- 「重申」（reiterated）通常不視為新資訊
