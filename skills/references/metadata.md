# metadata

**Type:** `dict` — 歷史資料的元數據

提供 `history` 方法返回資料的附加資訊。

## Dict Structure

| Key | Type | Description |
|-----|------|-------------|
| `symbol` | `str` | 股票代碼 |
| `currency` | `str` | 交易貨幣 |
| `exchange` | `str` | 交易所名稱 |
| `quoteType` | `str` | 資產類型 |
| `firstTradeDate` | `datetime` | 首次交易日期 |
| `regularMarketTime` | `datetime` | 正常市場時間 |
| `hasPrePostMarketData` | `bool` | 是否有盤前/盤後交易 |
| `timezone` | `str` | 交易所時區 |

## Field Details

### symbol（代碼）
- Yahoo Finance 使用的股票代碼
- 可能與交易所代碼略有不同

### currency（貨幣）
- 股票交易的結算貨幣
- 例如：USD, TWD, HKD, JPY

### exchange（交易所）
- 主要交易的交易所名稱
- 例如：NYSE, NASDAQ, TAIS

### quoteType（資產類型）

| 值 | 說明 |
|----|------|
| `EQUITY` | 股票 |
| `ETF` | 交易所交易基金 |
| `INDEX` | 指數 |
| `MUTUALFUND` | 共同基金 |
| `OPTION` | 期權 |
| `CURRENCY` | 貨幣 |

### firstTradeDate（首次交易日）
- 該股票首次在交易所交易的日期
- 可用於估算公司上市時間

### regularMarketTime（市場時間）
- 最近正常交易時段的時間戳
- UTC 格式

### hasPrePostMarketData（盤前/盤後）
- `True`：支持盤前和盤後交易
- `False`：僅支持正常交易時段

### timezone（時區）
- 交易所所在時區
- 例如：America/New_York, Asia/Taipei

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
info = ticker.history_metadata  # 或透過其他方式取得
print(f"Exchange: {info['exchange']}")
print(f"Currency: {info['currency']}")
print(f"First traded: {info['firstTradeDate']}")
```

## 與 info 的差異

| 特性 | `metadata` | `info` |
|------|------------|--------|
| 來源 | history() 方法 | info() 屬性 |
| 範圍 | 僅歷史資料相關 | 完整的公司資訊 |
| 即時性 | 靜態資訊 | 可能即時更新 |

## Notes

- 並非所有股票都有完整的 metadata
- timezone 可用於正確解讀時間戳
- 跨時區投資者需注意交易時間差異
