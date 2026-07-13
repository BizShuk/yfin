# sustainability

**Type:** `DataFrame` — ESG/永續性評分資料

提供公司的環境、社會及公司治理（ESG）相關評分。

## DataFrame Structure

欄位結構依資料可用性而異，通常包含：

| Column | Type | Description |
|--------|------|-------------|
| `Rating` | `str` | 總體 ESG 評級 |
| `E Score` | `float` | 環境評分 |
| `S Score` | `float` | 社會評分 |
| `G Score` | `float` | 公司治理評分 |

## Field Details

### Rating（總體評級）
- 綜合 ESG 表現的字母評級
- 例如：A, A+, B, BB, CCC

### E Score（環境評分）
- 環境績效指標
- 涵蓋：碳排放、能源效率、水資源使用、廢棄物管理

### S Score（社會評分）
- 社會責任指標
- 涵蓋：員工關係、社區影響、產品安全、供應鏈管理

### G Score（公司治理評分）
- 公司治理指標
- 涵蓋：董事會獨立性、透明度、薪酬政策、股東權利

## ESG 評分機構

| 機構 | 評分方式 |
|------|----------|
| MSCI | AAA-CCC |
| Sustainalytics | 風險評分（0-100） |
| ISS | Prime/Non-Prime |
| Refinitiv | D+-A+ |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
esg = ticker.sustainability
print(esg)
```

## 投資考量

```python
# 轉型風險：高 ESG 評分通常表示較低的長期風險
# 機會識別：符合永續發展趨勢的公司可能更有前景

# 不同投資者的關注點
# 價值投資者：更關注 G（治理）
# 成長投資者：可能忽略 ESG
# 永續投資者：對 E、S、G 都有要求
```

## 常見 ESG 指標

| 類別 | 指標 |
|------|------|
| 環境 | 碳排放量、再生能源使用、用水量 |
| 社會 | 員工多元化、工安紀錄、社區投資 |
| 治理 | 獨立董事比例、高管薪酬比、投票權結構 |

## Notes

- Yahoo Finance 的 ESG 資料來自第三方評分機構
- 並非所有股票都有完整的 ESG 資料
- 不同機構的評分標準可能有所不同
- ESG 評分並非投資建議
