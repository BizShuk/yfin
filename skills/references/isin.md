# isin

**Type:** `str` — 國際證券識別碼

ISIN（International Securities Identification Number）是證券的唯一識別碼。

## 基本格式

```
ISIN = 國碼(2) + 當地碼(9) + 校驗碼(1)
例如：US0378331005（Apple Inc.）
```

## 組成部分

| 組成 | 說明 |
|------|------|
| 國碼（前2位） | ISO 3166-1 alpha-2 國家代碼 |
| 當地碼（9位） | 各國證券交易所分配的代碼 |
| 校驗碼（1位） | 數學公式計算的校驗碼 |

## 常見國碼

| 國碼 | 國家/地區 |
|------|-----------|
| US | 美國 |
| TW | 台灣 |
| HK | 香港 |
| JP | 日本 |
| GB | 英國 |
| DE | 德國 |

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
isin = ticker.isin
print(f"ISIN: {isin}")
# 輸出：ISIN: US0378331005
```

## 與其他代碼的比較

| 代碼類型 | 格式 | 用途 |
|----------|------|------|
| ISIN | 12字元 | 國際通用識別 |
| CUSIP | 9字元 | 美國/加拿大 |
| SEDOL | 7字元 | 英國/愛爾蘭 |
| Ticker | 1-5字元 | 交易所代碼 |

## 用途

1. **國際結算**：跨境證券交易的核心識別碼
2. **研究分析**：區分不同市場的同名股票
3. **法規申報**：監管機構要求使用 ISIN

## Notes

- ISIN 是國際標準（ISO 6166）
- 並非所有股票都有 ISIN
- 美國股票的 ISIN 以 "US" 開頭
- 查詢 ISIN 可以到各大交易所網站或金融資料庫
