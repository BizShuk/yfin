# insider_roster

**Type:** `DataFrame` — 內部人名冊

列出公司的主要內部人及其基本資訊。

## DataFrame Structure

| Column | Type | Description |
|--------|------|-------------|
| `Name` | `str` | 內部人姓名 |
| `Title` | `str` | 職位/頭銜 |
| `火烧` | `str` | 匯報對象（直屬上司） |
| `Insider` | `str` | 是否為內部人（"Yes"/"No"） |

## Field Details

### Name（姓名）
- 內部人的完整姓名
- 通常為公司的高階主管或董事

### Title（職位）
- 該內部人的公司職位
- 例如：CEO, CFO, COO, Director, Vice President

### 火烧（匯報對象）
-該內部人的直屬上司
- 若為 CEO 或 Chairman，通常為空或 "None"
- 此欄位名稱可能為翻譯問題，應為 "Reports To"

### Insider（是否內部人）
- `"Yes"` 表示為公司內部人
- `"No"` 表示非內部人（可能是外部董事）

## 內部人類型

| 類型 | 說明 |
|------|------|
| 董事 | 董事會成員 |
| 高階主管 | CEO、CFO、COO 等管理層 |
| 員工 | 持有大量股票的員工 |
| 主要股東 | 持股超過 10% 的股東 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
roster = ticker.insider_roster
# 列出所有高階主管
executives = roster[roster['Title'].str.contains('President|CEO|CFO|COO', na=False)]
print(executives)
```

## 公司治理分析

```python
# 計算內部人董事比例
insider_directors = roster[(roster['Insider'] == 'Yes') & (roster['Title'].str.contains('Director', na=False))]
print(f"Insider Directors: {len(insider_directors)}")
```

## Notes

- 此資料主要用於了解公司管理團隊結構
- 並非所有內部人都會出現在此名冊
- Yahoo Finance 的欄位翻譯可能不正確
