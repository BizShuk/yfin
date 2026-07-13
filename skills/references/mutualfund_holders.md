# mutualfund_holders

**Type:** `DataFrame` — 共同基金持股資料

記錄持有該股票的所有共同基金及其持股狀況。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Holder` | `str` | 共同基金名稱 |
| `Shares` | `int` | 持股數量（股數） |
| `Date Reported` | `datetime` | 基金持股報告日期 |
| `% Held` | `float` | 持股比例（%） |
| `Value` | `float` | 持股市值（美元） |

## Field Details

### Holder（基金名稱）
- 共同基金的完整名稱
- 通常包含基金家族名稱（如 "Fidelity Contrafund"）
- 可能包含股份類別（如 "Class A", "Class I"）

### Shares（持股數量）
- 該基金持有的股數
- 共同基金通常持有大量股份

### Date Reported（報告日期）
- 基金的持股報告日期
- 多數基金每月公布持股明細

### % Held（持股比例）
- 該基金持股佔總發行股數的百分比
- 共同基金通常分散持股，單一基金很少超過 5-10%

### Value（持股市值）
- 持股數量 × 報告日股價
- 以美元計價

## 共同基金 vs 指數基金

| 特性 | 主動型共同基金 | 指數型基金/ETF |
|------|---------------|----------------|
| 經理人 | 主動選股 | 被動追蹤指數 |
| 費用率 | 通常較高 | 通常較低 |
| 持股變動 | 較頻繁 | 按指數調整 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("MSFT")
mf_holders = ticker.mutualfund_holders
print(mf_holders.head(10))
```

## 常見基金類型

| 類型 | 說明 |
|------|------|
| 成長型基金 | 追求資本利得 |
| 收益型基金 | 追求穩定股息 |
| 指數基金 | 複製大盤表現 |
| 產業基金 | 集中特定產業 |

## Notes

- 共同基金持股資料通常比機構投資人（institutional_holders）更全面
- 部分股票可能沒有共同基金持股資料
- 資料來源為基金的定期披露報告
