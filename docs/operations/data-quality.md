# Data Quality & Validation Guidelines

This guide provides comprehensive information about data quality expectations, validation strategies, and best practices for ensuring reliable data when using yfin.

## Table of Contents

1. [Data Quality Expectations](#data-quality-expectations)
2. [Validation Strategies](#validation-strategies)
3. [Data Quality Checks](#data-quality-checks)
4. [Handling Missing Data](#handling-missing-data)
5. [Data Consistency](#data-consistency)
6. [Quality Monitoring](#quality-monitoring)
7. [Best Practices](#best-practices)

## Data Quality Expectations

### Expected Data Availability

| Data Type | Availability | Quality Level | Notes |
|-----------|--------------|---------------|-------|
| **Quotes** | High (95%+) | Excellent | Generally available for all active stocks |
| **Historical Data** | High (90%+) | Excellent | Available for most stocks, limited by listing date |
| **Company Info** | High (95%+) | Good | Basic info only, detailed profiles not available |
| **Financials** | Medium (80%+) | Good | Available for most public companies |
| **Key Statistics** | Medium (75%+) | Good | May be limited for smaller companies |
| **Analysis** | Medium (70%+) | Variable | May be limited for smaller companies |
| **News** | Low (50%+) | Variable | Highly variable, may be empty for many stocks |

### Data Quality Indicators

#### High Quality Indicators
- ✅ Data fields are populated
- ✅ Timestamps are recent and valid
- ✅ Price data is reasonable (not zero or negative)
- ✅ Volume data is positive
- ✅ Currency codes are valid ISO-4217 codes
- ✅ Scaled decimal values have appropriate scales

#### Low Quality Indicators
- ❌ Missing required fields
- ❌ Zero or negative prices
- ❌ Negative volume
- ❌ Invalid timestamps
- ❌ Missing currency codes
- ❌ Inconsistent data formats

## Validation Strategies

### 1. Basic Data Validation

```go
func validateBarData(bar *facade.Bar) error {
    // Check for required fields
    if bar.Date == "" {
        return fmt.Errorf("missing date")
    }

    if bar.CurrencyCode == "" {
        return fmt.Errorf("missing currency code")
    }

    // Validate price data (facade.Bar already decoded from ScaledDecimal to float64)
    if bar.Open <= 0 || bar.High <= 0 || bar.Low <= 0 || bar.Close <= 0 {
        return fmt.Errorf("non-positive prices detected")
    }

    // Validate volume
    if bar.Volume < 0 {
        return fmt.Errorf("negative volume: %d", bar.Volume)
    }

    // Validate price relationships
    if err := validatePriceRelationships(bar); err != nil {
        return fmt.Errorf("invalid price relationships: %w", err)
    }

    return nil
}

func validatePriceRelationships(bar *facade.Bar) error {
    open, high, low, close := bar.Open, bar.High, bar.Low, bar.Close

    // High should be >= all other prices
    if high < open || high < low || high < close {
        return fmt.Errorf("high price is not the highest")
    }

    // Low should be <= all other prices
    if low > open || low > high || low > close {
        return fmt.Errorf("low price is not the lowest")
    }

    return nil
}
```

### 2. Quote Data Validation

```go
func validateQuoteData(quote *facade.Quote) error {
    // Check for required fields
    if quote.Symbol == "" {
        return fmt.Errorf("missing symbol")
    }

    if quote.Currency == "" {
        return fmt.Errorf("missing currency code")
    }

    if quote.EventTime.IsZero() {
        return fmt.Errorf("missing event time")
    }

    // facade.Quote.Price is float64; 0 means the market price was missing
    if quote.Price <= 0 {
        return fmt.Errorf("non-positive market price: %.2f", quote.Price)
    }

    return nil
}
```

### 3. Financial Data Validation

```go
func validateFinancialsData(financials *facade.FundamentalsSnapshot) error {
    if len(financials.Lines) == 0 {
        return fmt.Errorf("no financial lines found")
    }

    // Check for required metadata
    if financials.Symbol == "" {
        return fmt.Errorf("missing symbol")
    }

    // Validate each line item
    for i, line := range financials.Lines {
        if err := validateFinancialLine(&line); err != nil {
            return fmt.Errorf("invalid line %d: %w", i, err)
        }
    }

    return nil
}

func validateFinancialLine(line *facade.FundamentalsLine) error {
    if line.Key == "" {
        return fmt.Errorf("missing key")
    }

    if line.CurrencyCode == "" {
        return fmt.Errorf("missing currency code")
    }

    // Some financial metrics should be positive (Value is float64, decoded from ScaledDecimal)
    positiveMetrics := []string{"revenue", "net_income", "total_assets", "market_cap"}
    for _, metric := range positiveMetrics {
        if strings.Contains(strings.ToLower(line.Key), metric) && line.Value < 0 {
            return fmt.Errorf("negative value for positive metric %s: %.2f", line.Key, line.Value)
        }
    }

    return nil
}
```

## Data Quality Checks

### 1. Completeness Checks

```go
func checkDataCompleteness(data *StockData) []string {
    var issues []string
    
    // Check quote data
    if data.Quote == nil {
        issues = append(issues, "Missing quote data")
    } else {
        // facade.Quote.Price is float64; 0 means the market price was missing
        if data.Quote.Price <= 0 {
            issues = append(issues, "Missing or non-positive market price")
        }
        if data.Quote.Currency == "" {
            issues = append(issues, "Missing currency code")
        }
    }

    // Check company info
    if data.CompanyInfo == nil {
        issues = append(issues, "Missing company information")
    } else {
        if data.CompanyInfo.LongName == "" {
            issues = append(issues, "Missing company name")
        }
        if data.CompanyInfo.Exchange == "" {
            issues = append(issues, "Missing exchange information")
        }
    }

    // Check financials
    if data.Financials == nil {
        issues = append(issues, "Missing financial data")
    } else if len(data.Financials.Lines) == 0 {
        issues = append(issues, "Empty financial data")
    }

    // Check news
    if data.News == nil || len(data.News) == 0 {
        issues = append(issues, "No news articles available")
    }

    return issues
}
```

### 2. Consistency Checks

```go
func checkDataConsistency(data *StockData) []string {
    var issues []string
    
    // Check currency consistency
    if data.Quote != nil && data.CompanyInfo != nil {
        if data.Quote.Currency != data.CompanyInfo.Currency {
            issues = append(issues, fmt.Sprintf("Currency mismatch: quote=%s, company=%s",
                data.Quote.Currency, data.CompanyInfo.Currency))
        }
    }

    // Check symbol consistency (facade flattens Security.Symbol to Symbol)
    if data.Quote != nil && data.CompanyInfo != nil {
        if data.Quote.Symbol != data.CompanyInfo.Symbol {
            issues = append(issues, fmt.Sprintf("Symbol mismatch: quote=%s, company=%s",
                data.Quote.Symbol, data.CompanyInfo.Symbol))
        }
    }

    // Check timestamp consistency
    if data.Quote != nil && data.CompanyInfo != nil {
        // facade.CompanyInfo has no EventTime; use facade.Quote.EventTime + a sentinel
        // "as of" from the latest scrape/run id in your own wrapper.
        if !data.Quote.EventTime.IsZero() && time.Since(data.Quote.EventTime) > 7*24*time.Hour {
            issues = append(issues, fmt.Sprintf("Quote event time is stale: %s",
                data.Quote.EventTime.Format(time.RFC3339)))
        }
    }

    return issues
}
```

### 3. Reasonableness Checks

```go
func checkDataReasonableness(data *StockData) []string {
    var issues []string
    
    // Check price reasonableness (facade.Quote.Price is float64)
    if data.Quote != nil && data.Quote.Price > 0 {
        price := data.Quote.Price
        if price > 10000 {
            issues = append(issues, fmt.Sprintf("Unusually high price: %.2f", price))
        } else if price < 0.01 {
            issues = append(issues, fmt.Sprintf("Unusually low price: %.2f", price))
        }
    }

    // Check financial data reasonableness (facade.FundamentalsLine.Value is float64)
    if data.Financials != nil {
        for _, line := range data.Financials.Lines {
            if strings.Contains(strings.ToLower(line.Key), "revenue") {
                if line.Value > 1000000000000 { // 1 trillion
                    issues = append(issues, fmt.Sprintf("Extremely high revenue: %.2f", line.Value))
                }
            }
            if strings.Contains(strings.ToLower(line.Key), "market_cap") {
                if line.Value > 10000000000000 { // 10 trillion
                    issues = append(issues, fmt.Sprintf("Extremely high market cap: %.2f", line.Value))
                }
            }
        }
    }

    return issues
}
```

## Handling Missing Data

### 1. Graceful Degradation

```go
func processStockDataWithFallback(client *facade.Client, symbol string) *StockData {
    data := &StockData{Symbol: symbol}
    
    // Try to fetch quote (required)
    quote, err := client.FetchQuote(ctx, symbol, runID)
    if err != nil {
        data.Errors = append(data.Errors, fmt.Sprintf("Quote failed: %v", err))
        return data // Cannot continue without quote
    }
    data.Quote = quote
    
    // Try to fetch company info (optional)
    companyInfo, err := client.FetchCompanyInfo(ctx, symbol, runID)
    if err != nil {
        data.Warnings = append(data.Warnings, fmt.Sprintf("Company info failed: %v", err))
        // Continue without company info
    } else {
        data.CompanyInfo = companyInfo
    }
    
    // Try to fetch financials (optional)
    financials, err := client.ScrapeFinancials(ctx, symbol, runID)
    if err != nil {
        data.Warnings = append(data.Warnings, fmt.Sprintf("Financials failed: %v", err))
        // Continue without financials
    } else {
        data.Financials = financials
    }
    
    // Try to fetch news (optional)
    news, err := client.ScrapeNews(ctx, symbol, runID)
    if err != nil {
        data.Warnings = append(data.Warnings, fmt.Sprintf("News failed: %v", err))
        // Continue without news
    } else {
        data.News = news
    }
    
    return data
}
```

### 2. Data Imputation

```go
func imputeMissingData(data *StockData) {
    // Impute missing market price with historical average (facade.Quote.Price is float64)
    if data.Quote != nil && data.Quote.Price <= 0 {
        data.Warnings = append(data.Warnings, "Market price missing; using historical average")
        // data.Quote.Price = getAveragePrice(data.Symbol)  // populate via your own wrapper
    }

    // Impute missing company info
    if data.CompanyInfo == nil {
        data.CompanyInfo = &facade.CompanyInfo{
            Symbol:   data.Symbol,
            LongName: data.Symbol, // Use symbol as fallback
            Currency: "USD",       // Default currency
        }
        data.Warnings = append(data.Warnings, "Company info imputed with defaults")
    }
}
```

### 3. Data Quality Scoring

```go
type DataQualityScore struct {
    Overall    float64
    Completeness float64
    Consistency  float64
    Reasonableness float64
    Issues     []string
}

func calculateDataQualityScore(data *StockData) DataQualityScore {
    score := DataQualityScore{
        Issues: make([]string, 0),
    }
    
    // Calculate completeness score
    completenessIssues := checkDataCompleteness(data)
    score.Completeness = 1.0 - float64(len(completenessIssues))/4.0 // 4 main data types
    score.Issues = append(score.Issues, completenessIssues...)
    
    // Calculate consistency score
    consistencyIssues := checkDataConsistency(data)
    score.Consistency = 1.0 - float64(len(consistencyIssues))/3.0 // 3 consistency checks
    score.Issues = append(score.Issues, consistencyIssues...)
    
    // Calculate reasonableness score
    reasonablenessIssues := checkDataReasonableness(data)
    score.Reasonableness = 1.0 - float64(len(reasonablenessIssues))/5.0 // 5 reasonableness checks
    score.Issues = append(score.Issues, reasonablenessIssues...)
    
    // Calculate overall score
    score.Overall = (score.Completeness + score.Consistency + score.Reasonableness) / 3.0
    
    return score
}
```

## Data Consistency

### 1. Cross-Reference Validation

```go
func validateCrossReferences(data *StockData) []string {
    var issues []string
    
    // Validate symbol consistency across all data types
    symbols := make(map[string]bool)
    
    if data.Quote != nil {
        symbols[data.Quote.Security.Symbol] = true
    }
    
    if data.CompanyInfo != nil {
        symbols[data.CompanyInfo.Security.Symbol] = true
    }
    
    if data.Financials != nil {
        symbols[data.Financials.Security.Symbol] = true
    }
    
    if len(symbols) > 1 {
        issues = append(issues, "Symbol mismatch across data types")
    }
    
    // Validate currency consistency
    currencies := make(map[string]bool)
    
    if data.Quote != nil {
        currencies[data.Quote.CurrencyCode] = true
    }
    
    if data.CompanyInfo != nil {
        currencies[data.CompanyInfo.Currency] = true
    }
    
    if len(currencies) > 1 {
        issues = append(issues, "Currency mismatch across data types")
    }
    
    return issues
}
```

### 2. Temporal Consistency

```go
func validateTemporalConsistency(data *StockData) []string {
    var issues []string
    
    // Check if timestamps are reasonable
    now := time.Now()
    
    if data.Quote != nil {
        if data.Quote.EventTime.After(now) {
            issues = append(issues, "Quote timestamp is in the future")
        }
        
        if now.Sub(data.Quote.EventTime) > 7*24*time.Hour {
            issues = append(issues, "Quote data is older than 7 days")
        }
    }
    
    if data.CompanyInfo != nil {
        if data.CompanyInfo.EventTime.After(now) {
            issues = append(issues, "Company info timestamp is in the future")
        }
    }
    
    return issues
}
```

## Quality Monitoring

### 1. Quality Metrics

```go
type QualityMetrics struct {
    TotalRequests    int64
    SuccessfulRequests int64
    FailedRequests   int64
    DataQualityIssues int64
    AverageQualityScore float64
}

func (qm *QualityMetrics) RecordRequest(success bool, qualityScore float64) {
    qm.TotalRequests++
    
    if success {
        qm.SuccessfulRequests++
    } else {
        qm.FailedRequests++
    }
    
    if qualityScore < 0.8 {
        qm.DataQualityIssues++
    }
    
    // Update average quality score
    qm.AverageQualityScore = (qm.AverageQualityScore*float64(qm.TotalRequests-1) + qualityScore) / float64(qm.TotalRequests)
}
```

### 2. Quality Alerts

```go
func checkQualityAlerts(metrics QualityMetrics) []string {
    var alerts []string
    
    // Check success rate
    successRate := float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests)
    if successRate < 0.9 {
        alerts = append(alerts, fmt.Sprintf("Low success rate: %.2f%%", successRate*100))
    }
    
    // Check data quality
    if metrics.AverageQualityScore < 0.8 {
        alerts = append(alerts, fmt.Sprintf("Low data quality score: %.2f", metrics.AverageQualityScore))
    }
    
    // Check error rate
    errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
    if errorRate > 0.1 {
        alerts = append(alerts, fmt.Sprintf("High error rate: %.2f%%", errorRate*100))
    }
    
    return alerts
}
```

## Best Practices

### 1. Always Validate Data

```go
func processStockData(client *facade.Client, symbol string) (*StockData, error) {
    // Fetch data
    data := fetchStockData(client, symbol)
    
    // Validate data
    if err := validateStockData(data); err != nil {
        return nil, fmt.Errorf("data validation failed: %w", err)
    }
    
    // Check quality
    qualityScore := calculateDataQualityScore(data)
    if qualityScore.Overall < 0.7 {
        log.Printf("Warning: Low data quality score for %s: %.2f", symbol, qualityScore.Overall)
    }
    
    return data, nil
}
```

### 2. Implement Quality Gates

```go
func qualityGate(data *StockData) error {
    // Must have quote data
    if data.Quote == nil {
        return fmt.Errorf("quote data is required")
    }

    // Must have valid price (facade.Quote.Price is float64)
    if data.Quote.Price <= 0 {
        return fmt.Errorf("market price is required")
    }

    // Must have valid currency
    if data.Quote.Currency == "" {
        return fmt.Errorf("currency code is required")
    }

    // Must have recent data
    if !data.Quote.EventTime.IsZero() && time.Since(data.Quote.EventTime) > 24*time.Hour {
        return fmt.Errorf("data is too old")
    }

    return nil
}
```

### 3. Monitor Data Quality

```go
func monitorDataQuality(data *StockData) {
    qualityScore := calculateDataQualityScore(data)
    
    // Log quality issues
    if len(qualityScore.Issues) > 0 {
        log.Printf("Data quality issues for %s: %v", data.Symbol, qualityScore.Issues)
    }
    
    // Send metrics
    metrics.RecordDataQuality(data.Symbol, qualityScore.Overall)
    
    // Alert if quality is too low
    if qualityScore.Overall < 0.5 {
        alerting.SendAlert(fmt.Sprintf("Critical data quality issue for %s: %.2f", 
            data.Symbol, qualityScore.Overall))
    }
}
```

### 4. Handle Edge Cases

```go
func handleEdgeCases(data *StockData) {
    // Handle zero / non-positive prices (facade.Quote.Price is float64)
    if data.Quote != nil && data.Quote.Price <= 0 {
        log.Printf("Warning: Zero price detected for %s", data.Symbol)
        // Mark for manual review
    }

    // Handle stale data
    if data.Quote != nil && !data.Quote.EventTime.IsZero() {
        age := time.Since(data.Quote.EventTime)
        if age > 7*24*time.Hour {
            log.Printf("Warning: Stale data for %s: %v old", data.Symbol, age)
        }
    }
}
```

## Next Steps

- [API Reference](../api/reference.md) - Complete API documentation
- [Data Structures](../api/data-structures.md) - Detailed data structure guide
- [Complete Examples](../api/examples.md) - Working code examples
- [Error Handling Guide](error-handling.md) - Comprehensive error handling
- [Method Comparison](../comparisons/method-comparison.md) - Method comparison and use cases
