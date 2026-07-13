# Complete Working Examples

Working code samples that demonstrate the `yfin` SDK end-to-end: data fetching, formatting, error handling, batch processing, HTML report generation, data pipelines, and performance tuning. All examples import `github.com/bizshuk/yfin/facade` and consume plain structs (no `*norm.ScaledDecimal` math).

## Table of Contents

1. [Basic Data Fetching](#basic-data-fetching)
2. [Data Processing & Formatting](#data-processing--formatting)
3. [Error Handling & Retry Logic](#error-handling--retry-logic)
4. [Batch Processing](#batch-processing)
5. [HTML Report Generation](#html-report-generation)
6. [Data Pipeline Integration](#data-pipeline-integration)
7. [Performance Optimization](#performance-optimization)

## Basic Data Fetching

### Example 1: Fetch a Quote

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bizshuk/yfin/facade"
)

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("example-%d", time.Now().Unix())

    quote, err := client.FetchQuote(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }

    // Direct float64 — no scaling math.
    fmt.Printf("Symbol: %s\n", quote.Symbol)
    fmt.Printf("Price:  %.2f %s\n", quote.Price, quote.Currency)
    fmt.Printf("When:   %s\n", quote.EventTime.UTC().Format(time.RFC3339))
}
```

Note: when the regular market is closed, `quote.Price` will be `0` — use `quote.EventTime` to disambiguate from a live-but-free-fall tick.

### Example 2: Fetch Historical Bars

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bizshuk/yfin/facade"
)

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("historical-%d", time.Now().Unix())

    start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
    end := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

    bars, err := client.FetchDailyBars(ctx, "AAPL", start, end, true, runID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Fetched %d bars for %s (MIC=%s)\n", len(bars.Bars), bars.Symbol, bars.MIC)

    for _, bar := range bars.Bars {
        // Direct float64 access — no ScaledDecimal math.
        fmt.Printf("%s  O=%.2f H=%.2f L=%.2f C=%.2f V=%d\n",
            bar.Date,
            bar.Open, bar.High, bar.Low, bar.Close,
            bar.Volume)
    }
}
```

### Example 3: Fetch Company Information

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bizshuk/yfin/facade"
)

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("company-%d", time.Now().Unix())

    companyInfo, err := client.FetchCompanyInfo(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Symbol:        %s\n", companyInfo.Symbol)
    fmt.Printf("Long Name:     %s\n", companyInfo.LongName)
    fmt.Printf("Short Name:    %s\n", companyInfo.ShortName)
    fmt.Printf("Exchange:      %s\n", companyInfo.Exchange)
    fmt.Printf("Full Exchange: %s\n", companyInfo.FullExchangeName)
    fmt.Printf("Currency:      %s\n", companyInfo.Currency)
    fmt.Printf("Instrument:    %s\n", companyInfo.InstrumentType)
    fmt.Printf("Timezone:      %s\n", companyInfo.Timezone)

    fmt.Println("\nNote: For detailed company profiles, use alternative data sources.")
}
```

## Data Processing & Formatting

### Example 4: Financial Data Processing

Scrape financials and rank the lines by absolute value. No scaled-decimal math — every line value is already a `float64`.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"
    "sort"
    "time"

    "github.com/bizshuk/yfin/facade"
)

type FinancialMetric struct {
    Key   string
    Value float64
    Unit  string
}

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("financials-%d", time.Now().Unix())

    financials, err := client.ScrapeFinancials(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }

    var metrics []FinancialMetric
    for _, line := range financials.Lines {
        metrics = append(metrics, FinancialMetric{
            Key:   line.Key,
            Value: line.Value,
            Unit:  line.CurrencyCode,
        })
    }

    // Sort by absolute value, descending.
    sort.Slice(metrics, func(i, j int) bool {
        return math.Abs(metrics[i].Value) > math.Abs(metrics[j].Value)
    })

    fmt.Printf("Top 10 Financial Metrics for %s:\n", financials.Symbol)
    fmt.Println("=====================================")
    for i, metric := range metrics {
        if i >= 10 {
            break
        }
        fmt.Printf("%-20s: %15.2f %s\n", metric.Key, metric.Value, metric.Unit)
    }
}
```

### Example 5: News Data Processing

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sort"
    "time"

    "github.com/bizshuk/yfin/facade"
)

type NewsSummary struct {
    Title       string
    Summary     string
    PublishedAt time.Time
    Source      string
    URL         string
}

func main() {
    client := facade.NewClient()
    ctx := context.Background()
    runID := fmt.Sprintf("news-%d", time.Now().Unix())

    news, err := client.ScrapeNews(ctx, "AAPL", runID)
    if err != nil {
        log.Fatal(err)
    }
    if len(news) == 0 {
        fmt.Println("No news articles found")
        return
    }

    summaries := make([]NewsSummary, 0, len(news))
    for _, article := range news {
        summaries = append(summaries, NewsSummary{
            Title:       article.Title,
            Summary:     article.Summary,
            PublishedAt: article.PublishedAt,
            Source:      article.Source,
            URL:         article.URL,
        })
    }

    sort.Slice(summaries, func(i, j int) bool {
        return summaries[i].PublishedAt.After(summaries[j].PublishedAt)
    })

    fmt.Printf("Latest News for AAPL (%d articles):\n", len(summaries))
    fmt.Println("=====================================")
    for i, summary := range summaries {
        if i >= 5 {
            break
        }
        summaryText := summary.Summary
        if len(summaryText) > 100 {
            summaryText = summaryText[:100] + "..."
        }
        fmt.Printf("\n%d. %s\n", i+1, summary.Title)
        fmt.Printf("   Source: %s\n", summary.Source)
        fmt.Printf("   Published: %s\n", summary.PublishedAt.Format("2006-01-02 15:04"))
        fmt.Printf("   Summary: %s\n", summaryText)
        fmt.Printf("   URL: %s\n", summary.URL)
    }
}
```

## Error Handling & Retry Logic

### Example 6: Robust Error Handling

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/bizshuk/yfin/facade"
)

// classifyError buckets errors for retry policy.
func classifyError(err error) string {
    if err == nil {
        return "none"
    }
    s := err.Error()
    switch {
    case strings.Contains(s, "429") || strings.Contains(s, "rate limit"):
        return "rate_limit"
    case strings.Contains(s, "401") || strings.Contains(s, "Unauthorized"):
        return "authentication"
    case strings.Contains(s, "timeout") || strings.Contains(s, "connection"):
        return "network"
    case strings.Contains(s, "parse") || strings.Contains(s, "schema"):
        return "parse"
    default:
        return "unknown"
    }
}

// fetchWithRetry — bounded retries with exponential backoff and jitter.
func fetchWithRetry(client *facade.Client, symbol string, maxRetries int) (*facade.Quote, error) {
    ctx := context.Background()
    runID := fmt.Sprintf("retry-%d", time.Now().Unix())

    var lastErr error
    for i := 0; i < maxRetries; i++ {
        quote, err := client.FetchQuote(ctx, symbol, runID)
        if err == nil {
            return quote, nil
        }
        lastErr = err

        kind := classifyError(err)
        log.Printf("Attempt %d failed for %s: %s (%s)",
            i+1, symbol, err.Error(), kind)

        if i < maxRetries-1 {
            delay := time.Duration(i+1) * time.Second
            if kind == "rate_limit" {
                delay *= 2
            }
            log.Printf("Retrying in %v...", delay)
            time.Sleep(delay)
        }
    }
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}

func main() {
    client := facade.NewClient()
    symbol := "AAPL"

    quote, err := fetchWithRetry(client, symbol, 3)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%s: %.2f %s\n", quote.Symbol, quote.Price, quote.Currency)
}
```

### Example 7: Graceful Degradation

Collect everything available for a symbol without aborting the whole pipeline on the first failure. Note the plain-struct types: `*facade.Quote`, `*facade.CompanyInfo`, `*facade.FundamentalsSnapshot`, `[]facade.NewsItem`.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bizshuk/yfin/facade"
)

type StockData struct {
    Symbol      string
    Quote       *facade.Quote
    CompanyInfo *facade.CompanyInfo
    Financials  *facade.FundamentalsSnapshot
    News        []facade.NewsItem
    Errors      []string
}

func fetchStockData(client *facade.Client, symbol string) *StockData {
    ctx := context.Background()
    runID := fmt.Sprintf("stock-data-%d", time.Now().Unix())

    data := &StockData{Symbol: symbol, Errors: make([]string, 0)}

    if quote, err := client.FetchQuote(ctx, symbol, runID); err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Quote: %v", err))
    } else {
        data.Quote = quote
    }

    if companyInfo, err := client.FetchCompanyInfo(ctx, symbol, runID); err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Company Info: %v", err))
    } else {
        data.CompanyInfo = companyInfo
    }

    if financials, err := client.ScrapeFinancials(ctx, symbol, runID); err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Financials: %v", err))
    } else {
        data.Financials = financials
    }

    if news, err := client.ScrapeNews(ctx, symbol, runID); err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("News: %v", err))
    } else {
        data.News = news
    }

    return data
}

func main() {
    client := facade.NewClient()
    data := fetchStockData(client, "AAPL")

    fmt.Printf("Stock Data for %s:\n", data.Symbol)
    fmt.Println("==================")
    if data.Quote != nil {
        fmt.Printf("Price: %.2f %s\n", data.Quote.Price, data.Quote.Currency)
    }
    if data.CompanyInfo != nil {
        fmt.Printf("Company: %s (Exchange: %s)\n", data.CompanyInfo.LongName, data.CompanyInfo.Exchange)
    }
    if data.Financials != nil {
        fmt.Printf("Financial Metrics: %d available\n", len(data.Financials.Lines))
    }
    if len(data.News) > 0 {
        fmt.Printf("News Articles: %d available\n", len(data.News))
    }
    if len(data.Errors) > 0 {
        fmt.Println("\nErrors encountered:")
        for _, err := range data.Errors {
            fmt.Printf("  - %s\n", err)
        }
    }
}
```

## Batch Processing

### Example 8: Concurrent Batch Processing

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/bizshuk/yfin/facade"
)

type BatchResult struct {
    Symbol   string
    Quote    *facade.Quote
    Error    error
    Duration time.Duration
}

func processSymbol(client *facade.Client, symbol string) BatchResult {
    start := time.Now()
    ctx := context.Background()
    runID := fmt.Sprintf("batch-%d", time.Now().Unix())

    quote, err := client.FetchQuote(ctx, symbol, runID)
    return BatchResult{
        Symbol:   symbol,
        Quote:    quote,
        Error:    err,
        Duration: time.Since(start),
    }
}

func main() {
    client := facade.NewClient()

    symbols := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "META", "NVDA", "NFLX"}

    results := make(chan BatchResult, len(symbols))
    var wg sync.WaitGroup

    // Bound concurrency to stay below Yahoo's per-IP QPS.
    semaphore := make(chan struct{}, 3)

    for _, symbol := range symbols {
        wg.Add(1)
        go func(sym string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            results <- processSymbol(client, sym)
        }(symbol)
    }

    go func() {
        wg.Wait()
        close(results)
    }()

    var successful, failed int
    var totalDuration time.Duration

    fmt.Printf("Processing %d symbols...\n", len(symbols))
    fmt.Println("================================")
    for result := range results {
        totalDuration += result.Duration
        if result.Error != nil {
            failed++
            fmt.Printf("FAIL %s: %v (%.2fs)\n", result.Symbol, result.Error, result.Duration.Seconds())
            continue
        }
        successful++
        fmt.Printf("OK %s: %.2f %s (%.2fs)\n",
            result.Symbol, result.Quote.Price, result.Quote.Currency, result.Duration.Seconds())
    }

    fmt.Println("\nBatch Processing Summary:")
    fmt.Printf("Successful: %d\n", successful)
    fmt.Printf("Failed:     %d\n", failed)
    fmt.Printf("Success Rate: %.1f%%\n", float64(successful)/float64(len(symbols))*100)
    fmt.Printf("Average Duration: %.2fs\n", totalDuration.Seconds()/float64(len(symbols)))
}
```

## HTML Report Generation

### Example 9: Generate an HTML Report

Templates now read `quote.Price` / `line.Value` directly — no `div`/`pow` helpers needed.

```go
package main

import (
    "context"
    "fmt"
    "html/template"
    "log"
    "os"
    "time"

    "github.com/bizshuk/yfin/facade"
)

type ReportData struct {
    Symbol      string
    Quote       *facade.Quote
    CompanyInfo *facade.CompanyInfo
    Financials  *facade.FundamentalsSnapshot
    News        []facade.NewsItem
    GeneratedAt time.Time
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Stock Report - {{.Symbol}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .metric { display: inline-block; margin: 10px; padding: 10px; background-color: #e8f4f8; border-radius: 3px; }
        .news-item { border-bottom: 1px solid #eee; padding: 10px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Stock Report: {{.Symbol}}</h1>
        <p>Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
    </div>

    {{if .Quote}}
    <div class="section">
        <h2>Current Quote</h2>
        <div class="metric"><strong>Price:</strong> {{printf "%.2f" .Quote.Price}} {{.Quote.Currency}}</div>
        <div class="metric"><strong>Event Time:</strong> {{.Quote.EventTime.Format "2006-01-02 15:04:05"}}</div>
    </div>
    {{end}}

    {{if .CompanyInfo}}
    <div class="section">
        <h2>Company Information</h2>
        <div class="metric"><strong>Name:</strong> {{.CompanyInfo.LongName}}</div>
        <div class="metric"><strong>Exchange:</strong> {{.CompanyInfo.Exchange}}</div>
        <div class="metric"><strong>Currency:</strong> {{.CompanyInfo.Currency}}</div>
        <div class="metric"><strong>Instrument Type:</strong> {{.CompanyInfo.InstrumentType}}</div>
    </div>
    {{end}}

    {{if .Financials}}
    <div class="section">
        <h2>Financial Metrics</h2>
        {{range .Financials.Lines}}
        <div class="metric"><strong>{{.Key}}:</strong> {{printf "%.2f" .Value}} {{.CurrencyCode}}</div>
        {{end}}
    </div>
    {{end}}

    {{if .News}}
    <div class="section">
        <h2>Recent News ({{len .News}} articles)</h2>
        {{range .News}}
        <div class="news-item">
            <h3>{{.Title}}</h3>
            <p><strong>Source:</strong> {{.Source}} | <strong>Published:</strong> {{.PublishedAt.Format "2006-01-02 15:04"}}</p>
            <p>{{.Summary}}</p>
            <p><a href="{{.URL}}" target="_blank">Read more</a></p>
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>
`

func generateReport(client *facade.Client, symbol string) error {
    ctx := context.Background()
    runID := fmt.Sprintf("report-%d", time.Now().Unix())

    data := &ReportData{Symbol: symbol, GeneratedAt: time.Now()}

    if quote, err := client.FetchQuote(ctx, symbol, runID); err == nil {
        data.Quote = quote
    }
    if companyInfo, err := client.FetchCompanyInfo(ctx, symbol, runID); err == nil {
        data.CompanyInfo = companyInfo
    }
    if financials, err := client.ScrapeFinancials(ctx, symbol, runID); err == nil {
        data.Financials = financials
    }
    if news, err := client.ScrapeNews(ctx, symbol, runID); err == nil {
        data.News = news
    }

    tmpl, err := template.New("report").Parse(htmlTemplate)
    if err != nil {
        return err
    }
    filename := fmt.Sprintf("stock_report_%s_%s.html", symbol, time.Now().Format("20060102_150405"))
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    return tmpl.Execute(file, data)
}

func main() {
    client := facade.NewClient()
    symbol := "AAPL"
    fmt.Printf("Generating report for %s...\n", symbol)
    if err := generateReport(client, symbol); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Report generated successfully!\n")
}
```

## Data Pipeline Integration

### Example 10: Data Pipeline with Validation

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/bizshuk/yfin/facade"
)

type DataPipeline struct {
    client *facade.Client
    runID  string
}

type PipelineResult struct {
    Symbol      string
    Quote       *facade.Quote
    CompanyInfo *facade.CompanyInfo
    Financials  *facade.FundamentalsSnapshot
    News        []facade.NewsItem
    Errors      []string
    Warnings    []string
}

func NewDataPipeline() *DataPipeline {
    return &DataPipeline{
        client: facade.NewClient(),
        runID:  fmt.Sprintf("pipeline-%d", time.Now().Unix()),
    }
}

func (p *DataPipeline) ProcessSymbol(symbol string) *PipelineResult {
    ctx := context.Background()
    result := &PipelineResult{
        Symbol:   symbol,
        Errors:   make([]string, 0),
        Warnings: make([]string, 0),
    }

    // Quote — required.
    if quote, err := p.client.FetchQuote(ctx, symbol, p.runID); err != nil {
        result.Errors = append(result.Errors, fmt.Sprintf("Quote fetch failed: %v", err))
        return result
    } else {
        result.Quote = quote
        if quote.Price == 0 {
            result.Warnings = append(result.Warnings, "Market closed — price is 0")
        }
    }

    // Company info — optional.
    if companyInfo, err := p.client.FetchCompanyInfo(ctx, symbol, p.runID); err != nil {
        result.Warnings = append(result.Warnings, fmt.Sprintf("Company info unavailable: %v", err))
    } else {
        result.CompanyInfo = companyInfo
    }

    // Financials — optional.
    if financials, err := p.client.ScrapeFinancials(ctx, symbol, p.runID); err != nil {
        result.Warnings = append(result.Warnings, fmt.Sprintf("Financials unavailable: %v", err))
    } else {
        result.Financials = financials
        if len(financials.Lines) == 0 {
            result.Warnings = append(result.Warnings, "No financial metrics available")
        }
    }

    // News — optional.
    if news, err := p.client.ScrapeNews(ctx, symbol, p.runID); err != nil {
        result.Warnings = append(result.Warnings, fmt.Sprintf("News unavailable: %v", err))
    } else {
        result.News = news
        if len(news) == 0 {
            result.Warnings = append(result.Warnings, "No news articles available")
        }
    }

    return result
}

func (p *DataPipeline) ProcessSymbols(symbols []string) []*PipelineResult {
    results := make([]*PipelineResult, 0, len(symbols))
    for i, symbol := range symbols {
        fmt.Printf("[%d/%d] Processing %s...", i+1, len(symbols), symbol)
        result := p.ProcessSymbol(symbol)
        results = append(results, result)
        if len(result.Errors) > 0 {
            fmt.Printf(" FAIL (%d errors)\n", len(result.Errors))
        } else {
            fmt.Printf(" OK   (%d warnings)\n", len(result.Warnings))
        }
        if i < len(symbols)-1 {
            time.Sleep(1 * time.Second)
        }
    }
    return results
}

func main() {
    pipeline := NewDataPipeline()
    symbols := []string{"AAPL", "MSFT", "GOOGL"}

    fmt.Printf("Starting data pipeline for %d symbols...\n", len(symbols))
    fmt.Println("==========================================")

    results := pipeline.ProcessSymbols(symbols)

    fmt.Println("\nPipeline Summary:")
    fmt.Println("=================")
    successful := 0
    totalWarnings := 0
    for _, result := range results {
        if len(result.Errors) == 0 {
            successful++
        }
        totalWarnings += len(result.Warnings)
        if len(result.Errors) > 0 {
            fmt.Printf("%s: FAILED  (%d errors)\n", result.Symbol, len(result.Errors))
        } else {
            fmt.Printf("%s: SUCCESS (%d warnings)\n", result.Symbol, len(result.Warnings))
        }
    }
    fmt.Printf("\nOverall: %d/%d successful (%.1f%%)\n",
        successful, len(symbols), float64(successful)/float64(len(symbols))*100)
    fmt.Printf("Total warnings: %d\n", totalWarnings)
}
```

## Performance Optimization

### Example 11: Optimized Batch Processing

Tune concurrency (semaphore) and per-request pacing (sleep) to land under Yahoo's undocumented per-IP QPS without dropping throughput.

```go
package main

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/bizshuk/yfin/facade"
)

type BatchResult struct {
    Symbol   string
    Quote    *facade.Quote
    Error    error
    Duration time.Duration
}

type OptimizedBatchProcessor struct {
    client      *facade.Client
    concurrency int
    rateLimit   time.Duration
}

func NewOptimizedBatchProcessor(concurrency int, rateLimit time.Duration) *OptimizedBatchProcessor {
    return &OptimizedBatchProcessor{
        client:      facade.NewClient(),
        concurrency: concurrency,
        rateLimit:   rateLimit,
    }
}

func (p *OptimizedBatchProcessor) ProcessSymbols(symbols []string) []BatchResult {
    results := make([]BatchResult, len(symbols))
    semaphore := make(chan struct{}, p.concurrency)
    var wg sync.WaitGroup

    for i, symbol := range symbols {
        wg.Add(1)
        go func(index int, sym string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            start := time.Now()
            ctx := context.Background()
            runID := fmt.Sprintf("optimized-%d", time.Now().Unix())
            quote, err := p.client.FetchQuote(ctx, sym, runID)

            results[index] = BatchResult{
                Symbol:   sym,
                Quote:    quote,
                Error:    err,
                Duration: time.Since(start),
            }
            time.Sleep(p.rateLimit)
        }(i, symbol)
    }
    wg.Wait()
    return results
}

func main() {
    processor := NewOptimizedBatchProcessor(3, 500*time.Millisecond)

    symbols := []string{
        "AAPL", "MSFT", "GOOGL", "AMZN", "TSLA", "META", "NVDA", "NFLX",
        "ORCL", "CRM", "ADBE", "INTC", "CSCO", "IBM", "AMD", "QCOM",
    }

    fmt.Printf("Processing %d symbols with optimized batch processor...\n", len(symbols))
    fmt.Printf("Concurrency: %d, Rate limit: %v\n", processor.concurrency, processor.rateLimit)
    fmt.Println("=====================================================")

    start := time.Now()
    results := processor.ProcessSymbols(symbols)
    totalDuration := time.Since(start)

    successful := 0
    var totalRequestTime time.Duration
    for _, result := range results {
        if result.Error == nil {
            successful++
        }
        totalRequestTime += result.Duration
    }

    fmt.Printf("\nOptimized Batch Processing Results:\n")
    fmt.Printf("Total time:           %.2fs\n", totalDuration.Seconds())
    fmt.Printf("Successful:           %d/%d (%.1f%%)\n",
        successful, len(symbols), float64(successful)/float64(len(symbols))*100)
    fmt.Printf("Average request time: %.2fs\n", totalRequestTime.Seconds()/float64(len(symbols)))
    fmt.Printf("Throughput:           %.2f requests/second\n", float64(len(symbols))/totalDuration.Seconds())
}
```

## Next Steps

- [API Reference](reference.md) — complete API documentation.
- [Migration Guide](../integrations/migration-guide.md) — moving from Python `yfinance` or earlier SDK surfaces.
