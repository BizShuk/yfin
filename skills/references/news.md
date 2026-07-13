# news

**Type:** `list[dict]` — 新聞文章列表

提供與該股票相關的最新新聞文章。

## List of Dict Structure

每篇文章為一個 dict，包含以下欄位：

| Key | Type | Description |
|-----|------|-------------|
| `id` | `str` | 文章唯一識別碼 |
| `content` | `str` | 文章全文或摘要內容 |
| `title` | `str` | 文章標題 |
| `pubDate` | `str` | 發布日期（ISO 8601 格式） |
| `link` | `str` | 文章原始連結 URL |
| `source` | `str` | 新聞來源媒體名稱 |

## Field Details

### id（文章 ID）
- 新聞系統的內部識別碼
- 可用於去重或追蹤特定文章

### content（內容）
- 文章的正文或摘要
- 長度可能受限於 Yahoo Finance 的 API 限制

### title（標題）
- 新聞文章的標題
- 通常是點擊率最高的部分

### pubDate（發布日期）
- ISO 8601 格式的日期時間字串
- 例如："2025-05-27T10:30:00Z"
- 可能為 UTC 或當地時間

### link（連結）
- 原始文章的 URL
- 可能需要登入或付費才能閱讀完整內容

### source（來源）
- 新聞媒體或內容提供者的名稱
- 例如："Reuters", "Bloomberg", "CNBC"

## Example

```python
import yfinance as yf
ticker = yf.Ticker("AAPL")
news = ticker.news
for article in news[:5]:
    print(f"[{article['pubDate']}] {article['title']}")
    print(f"Source: {article['source']}")
    print(f"Link: {article['link']}")
    print()
```

## 常見新聞來源

| 來源 | 類型 |
|------|------|
| Reuters | 通訊社 |
| Bloomberg | 財經媒體 |
| CNBC | 電視台 |
| WSJ | 報紙 |
| Motley Fool | 投資網站 |

## 資料時間範圍

- 通常包含最近 30 天的新聞
- 熱門股票可能包含更多文章
- 時間範圍可透過 `days` 參數控制

## Notes

- Yahoo Finance 的新聞可能來自多個來源
- 部分付費內容可能只有摘要
- 新聞列表為實時更新，可能與取得時略有差異
