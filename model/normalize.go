// normalize.go — `Normalize*` conversion functions from raw Yahoo Finance
// response shapes (now in `model/yahoo_raw.go`) into `Normalized*` data
// types. Originally split across svc/norm/{bars,quotes,fundamentals,
// market_data,company_info}.go; consolidated here so any layer can
// construct normalized data without pulling in svc/norm.
//
// Function bodies retain the original logic verbatim (currency-aware
// ScaledDecimal conversion, MIC inference, UTC timestamp normalization).
// All inputs are `model.*` types — no `svc/yahoo` import, no import
// cycle. svc/yahoo's decoder returns `*model.ChartResponse` /
// `*model.QuoteResponse` / `*model.FundamentalsResponse` directly.

package model

import (
	"fmt"
	"time"
)

// NormalizeBars converts Yahoo Finance bars to normalized bars.
func NormalizeBars(bars []ChartBar, meta *ChartMeta, runID string) (*NormalizedBarBatch, error) {
	if len(bars) == 0 {
		return nil, fmt.Errorf("no bars to normalize")
	}
	if meta == nil {
		return nil, fmt.Errorf("missing metadata")
	}

	security := CreateSecurity(meta.Symbol, meta.ExchangeName, meta.ExchangeName)
	if err := ValidateSecurity(security); err != nil {
		return nil, fmt.Errorf("invalid security: %w", err)
	}

	isAdjusted := false
	adjustmentPolicyID := "raw"
	for _, bar := range bars {
		if bar.AdjClose != nil {
			isAdjusted = true
			adjustmentPolicyID = "split_dividend"
			break
		}
	}

	scale := GetScaleForCurrency(meta.Currency)

	normalizedBars := make([]NormalizedBar, 0, len(bars))
	ingestTime := time.Now().UTC()

	for _, bar := range bars {
		normalizedBar, err := normalizeBar(bar, meta.Currency, scale, isAdjusted, adjustmentPolicyID, ingestTime)
		if err != nil {
			continue
		}
		normalizedBars = append(normalizedBars, normalizedBar)
	}

	if len(normalizedBars) == 0 {
		return nil, fmt.Errorf("no valid bars after normalization")
	}

	metaData := Meta{
		RunID:         runID,
		Source:        "yfinance-go",
		Producer:      "local",
		SchemaVersion: "ampy.bars.v1:1.0.0",
	}

	return &NormalizedBarBatch{
		Security: security,
		Bars:     normalizedBars,
		Meta:     metaData,
	}, nil
}

func normalizeBar(bar ChartBar, currency string, scale int, isAdjusted bool, adjustmentPolicyID string, now time.Time) (NormalizedBar, error) {
	start, end, eventTime := ToUTCDayBoundaries(bar.Timestamp)

	closePrice := bar.Close
	if isAdjusted && bar.AdjClose != nil {
		closePrice = *bar.AdjClose
	}

	open, err := ToScaledDecimalWithCurrency(bar.Open, currency)
	if err != nil {
		return NormalizedBar{}, fmt.Errorf("invalid open price: %w", err)
	}
	high, err := ToScaledDecimalWithCurrency(bar.High, currency)
	if err != nil {
		return NormalizedBar{}, fmt.Errorf("invalid high price: %w", err)
	}
	low, err := ToScaledDecimalWithCurrency(bar.Low, currency)
	if err != nil {
		return NormalizedBar{}, fmt.Errorf("invalid low price: %w", err)
	}
	closePriceScaled, err := ToScaledDecimalWithCurrency(closePrice, currency)
	if err != nil {
		return NormalizedBar{}, fmt.Errorf("invalid close price: %w", err)
	}

	return NormalizedBar{
		Start:              start,
		End:                end,
		Open:               open,
		High:               high,
		Low:                low,
		Close:              closePriceScaled,
		Volume:             bar.Volume,
		Adjusted:           isAdjusted,
		AdjustmentPolicyID: adjustmentPolicyID,
		CurrencyCode:       currency,
		EventTime:          eventTime,
		IngestTime:         now,
		AsOf:               eventTime,
	}, nil
}

// NormalizeQuote converts a Yahoo Finance quote to a normalized quote.
func NormalizeQuote(quote RawQuote, runID string) (*NormalizedQuote, error) {
	if quote.Symbol == "" {
		return nil, fmt.Errorf("missing symbol")
	}
	if quote.Currency == "" {
		return nil, fmt.Errorf("missing currency")
	}

	security := CreateSecurity(quote.Symbol, quote.Exchange, quote.FullExchangeName)
	if err := ValidateSecurity(security); err != nil {
		return nil, fmt.Errorf("invalid security: %w", err)
	}

	scale := GetScaleForCurrency(quote.Currency)
	eventTime := time.Now().UTC()

	var bid, ask *ScaledDecimal
	if quote.Bid != nil {
		bidScaled, err := ToScaledDecimal(*quote.Bid, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid bid price: %w", err)
		}
		bid = &bidScaled
	}
	if quote.Ask != nil {
		askScaled, err := ToScaledDecimal(*quote.Ask, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid ask price: %w", err)
		}
		ask = &askScaled
	}

	var regularMarketPrice, regularMarketHigh, regularMarketLow *ScaledDecimal
	if quote.RegularMarketPrice != nil {
		priceScaled, err := ToScaledDecimal(*quote.RegularMarketPrice, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid regular market price: %w", err)
		}
		regularMarketPrice = &priceScaled
	}
	if quote.RegularMarketDayHigh != nil {
		highScaled, err := ToScaledDecimal(*quote.RegularMarketDayHigh, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid regular market high: %w", err)
		}
		regularMarketHigh = &highScaled
	}
	if quote.RegularMarketDayLow != nil {
		lowScaled, err := ToScaledDecimal(*quote.RegularMarketDayLow, scale)
		if err != nil {
			return nil, fmt.Errorf("invalid regular market low: %w", err)
		}
		regularMarketLow = &lowScaled
	}

	venue := ""
	if quote.Exchange != "" {
		venue = InferMIC(quote.Exchange, "")
		if venue == "" {
			venue = quote.Exchange
		}
	}

	meta := Meta{
		RunID:         runID,
		Source:        "yfinance-go",
		Producer:      "local",
		SchemaVersion: "ampy.ticks.v1:1.0.0",
	}

	return &NormalizedQuote{
		Security:            security,
		Type:                "QUOTE",
		Bid:                 bid,
		BidSize:             quote.BidSize,
		Ask:                 ask,
		AskSize:             quote.AskSize,
		RegularMarketPrice:  regularMarketPrice,
		RegularMarketHigh:   regularMarketHigh,
		RegularMarketLow:    regularMarketLow,
		RegularMarketVolume: quote.RegularMarketVolume,
		Venue:               venue,
		CurrencyCode:        quote.Currency,
		EventTime:           eventTime,
		IngestTime:          eventTime,
		Meta:                meta,
	}, nil
}

// NormalizeFundamentals converts Yahoo Finance fundamentals to normalized fundamentals.
func NormalizeFundamentals(fundamentals *Fundamentals, symbol, runID string) (*NormalizedFundamentalsSnapshot, error) {
	if fundamentals == nil {
		return nil, fmt.Errorf("no fundamentals data")
	}

	security := Security{
		Symbol: symbol,
		MIC:    "",
	}
	if err := ValidateSecurity(security); err != nil {
		return nil, fmt.Errorf("invalid security: %w", err)
	}

	lines := make([]NormalizedFundamentalsLine, 0)

	for _, stmt := range fundamentals.IncomeStatements {
		stmtLines, err := normalizeIncomeStatement(stmt)
		if err != nil {
			continue
		}
		lines = append(lines, stmtLines...)
	}
	for _, sheet := range fundamentals.BalanceSheets {
		sheetLines, err := normalizeBalanceSheet(sheet)
		if err != nil {
			continue
		}
		lines = append(lines, sheetLines...)
	}
	for _, stmt := range fundamentals.CashflowStatements {
		stmtLines, err := normalizeCashflowStatement(stmt)
		if err != nil {
			continue
		}
		lines = append(lines, stmtLines...)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("no valid fundamentals lines found")
	}

	meta := Meta{
		RunID:         runID,
		Source:        "yfinance-go",
		Producer:      "local",
		SchemaVersion: "ampy.fundamentals.v1:1.0.0",
	}

	return &NormalizedFundamentalsSnapshot{
		Security: security,
		Lines:    lines,
		Source:   "yfinance",
		AsOf:     time.Date(2025, 8, 1, 0, 0, 0, 0, time.UTC),
		Meta:     meta,
	}, nil
}

func normalizeIncomeStatement(stmt IncomeStatement) ([]NormalizedFundamentalsLine, error) {
	lines := make([]NormalizedFundamentalsLine, 0)
	periodStart, periodEnd := convertDateToPeriod(stmt.EndDate)

	if stmt.TotalRevenue != nil && stmt.TotalRevenue.Raw != nil {
		if line, err := createFundamentalsLine("revenue", *stmt.TotalRevenue.Raw, "USD", periodStart, periodEnd); err == nil {
			lines = append(lines, line)
		}
	}
	if stmt.NetIncome != nil && stmt.NetIncome.Raw != nil {
		if line, err := createFundamentalsLine("net_income", *stmt.NetIncome.Raw, "USD", periodStart, periodEnd); err == nil {
			lines = append(lines, line)
		}
	}
	if stmt.EPS != nil && stmt.EPS.Raw != nil {
		if line, err := createFundamentalsLine("eps_basic", *stmt.EPS.Raw, "USD", periodStart, periodEnd); err == nil {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func normalizeBalanceSheet(sheet BalanceSheet) ([]NormalizedFundamentalsLine, error) {
	lines := make([]NormalizedFundamentalsLine, 0)
	periodStart, periodEnd := convertDateToPeriod(sheet.EndDate)

	if sheet.TotalAssets != nil && sheet.TotalAssets.Raw != nil {
		if line, err := createFundamentalsLine("total_assets", *sheet.TotalAssets.Raw, "USD", periodStart, periodEnd); err == nil {
			lines = append(lines, line)
		}
	}
	if sheet.TotalLiab != nil && sheet.TotalLiab.Raw != nil {
		if line, err := createFundamentalsLine("total_liabilities", *sheet.TotalLiab.Raw, "USD", periodStart, periodEnd); err == nil {
			lines = append(lines, line)
		}
	}
	if sheet.TotalStockholderEquity != nil && sheet.TotalStockholderEquity.Raw != nil {
		if line, err := createFundamentalsLine("total_equity", *sheet.TotalStockholderEquity.Raw, "USD", periodStart, periodEnd); err == nil {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func normalizeCashflowStatement(stmt CashflowStatement) ([]NormalizedFundamentalsLine, error) {
	lines := make([]NormalizedFundamentalsLine, 0)
	periodStart, periodEnd := convertDateToPeriod(stmt.EndDate)

	if stmt.NetIncome != nil && stmt.NetIncome.Raw != nil {
		if line, err := createFundamentalsLine("net_income", *stmt.NetIncome.Raw, "USD", periodStart, periodEnd); err == nil {
			lines = append(lines, line)
		}
	}
	if stmt.TotalCashFromOperatingActivities != nil && stmt.TotalCashFromOperatingActivities.Raw != nil {
		if line, err := createFundamentalsLine("operating_cashflow", *stmt.TotalCashFromOperatingActivities.Raw, "USD", periodStart, periodEnd); err == nil {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func createFundamentalsLine(key string, value int64, currency string, periodStart, periodEnd time.Time) (NormalizedFundamentalsLine, error) {
	if key == "" {
		return NormalizedFundamentalsLine{}, fmt.Errorf("key cannot be empty")
	}
	if currency == "" {
		return NormalizedFundamentalsLine{}, fmt.Errorf("currency cannot be empty")
	}
	if periodStart.After(periodEnd) {
		return NormalizedFundamentalsLine{}, fmt.Errorf("period start cannot be after period end")
	}

	var scaled ScaledDecimal
	var err error
	if key == "eps_basic" {
		scaled = ScaledDecimal{
			Scaled: value,
			Scale:  2,
		}
	} else {
		scaled, err = ToScaledDecimal(float64(value), 2)
		if err != nil {
			return NormalizedFundamentalsLine{}, fmt.Errorf("invalid value for %s: %w", key, err)
		}
	}

	return NormalizedFundamentalsLine{
		Key:          key,
		Value:        scaled,
		CurrencyCode: currency,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}, nil
}

func convertDateToPeriod(dateValue DateValue) (periodStart, periodEnd time.Time) {
	if dateValue.Raw != 0 {
		periodEnd = time.Unix(dateValue.Raw, 0).UTC()
		periodStart = periodEnd.AddDate(0, -3, 0)
	} else {
		periodEnd = time.Now().UTC()
		periodStart = periodEnd.AddDate(0, -3, 0)
	}
	return periodStart, periodEnd
}

// NormalizeMarketData normalizes comprehensive market data from chart metadata.
func NormalizeMarketData(meta *ChartMeta, runID string) (*NormalizedMarketData, error) {
	if meta == nil {
		return nil, fmt.Errorf("metadata is nil")
	}

	security := Security{
		Symbol: meta.Symbol,
		MIC:    InferMIC(meta.ExchangeName, ""),
	}

	var regularMarketTime *time.Time
	if meta.RegularMarketTime != 0 {
		rmt := time.Unix(meta.RegularMarketTime, 0).UTC()
		regularMarketTime = &rmt
	}

	marketData := &NormalizedMarketData{
		Security:             security,
		RegularMarketPrice:   ToScaledDecimalPtr(meta.RegularMarketPrice, meta.Currency),
		RegularMarketHigh:    ToScaledDecimalPtr(meta.RegularMarketDayHigh, meta.Currency),
		RegularMarketLow:     ToScaledDecimalPtr(meta.RegularMarketDayLow, meta.Currency),
		RegularMarketVolume:  meta.RegularMarketVolume,
		FiftyTwoWeekHigh:     ToScaledDecimalPtr(meta.FiftyTwoWeekHigh, meta.Currency),
		FiftyTwoWeekLow:      ToScaledDecimalPtr(meta.FiftyTwoWeekLow, meta.Currency),
		PreviousClose:        ToScaledDecimalPtr(meta.PreviousClose, meta.Currency),
		ChartPreviousClose:   ToScaledDecimalPtr(meta.ChartPreviousClose, meta.Currency),
		RegularMarketTime:    regularMarketTime,
		HasPrePostMarketData: meta.HasPrePostMarketData,
		EventTime:            time.Now().UTC(),
		IngestTime:           time.Now().UTC(),
		Meta: Meta{
			RunID:         runID,
			Source:        "yahoo",
			Producer:      "yfinance-go",
			SchemaVersion: "1.0",
		},
	}

	return marketData, nil
}

// NormalizeCompanyInfo normalizes company information from chart metadata.
func NormalizeCompanyInfo(meta *ChartMeta, runID string) (*NormalizedCompanyInfo, error) {
	if meta == nil {
		return nil, fmt.Errorf("metadata is nil")
	}

	security := Security{
		Symbol: meta.Symbol,
		MIC:    InferMIC(meta.ExchangeName, ""),
	}

	var firstTradeDate *time.Time
	if meta.FirstTradeDate != 0 {
		ftd := time.Unix(meta.FirstTradeDate, 0).UTC()
		firstTradeDate = &ftd
	}

	companyInfo := &NormalizedCompanyInfo{
		Security:         security,
		LongName:         meta.LongName,
		ShortName:        meta.ShortName,
		Exchange:         meta.ExchangeName,
		FullExchangeName: meta.FullExchangeName,
		Currency:         meta.Currency,
		InstrumentType:   meta.InstrumentType,
		FirstTradeDate:   firstTradeDate,
		Timezone:         meta.Timezone,
		ExchangeTimezone: meta.ExchangeTimezoneName,
		EventTime:        time.Now().UTC(),
		IngestTime:       time.Now().UTC(),
		Meta: Meta{
			RunID:         runID,
			Source:        "yahoo",
			Producer:      "yfinance-go",
			SchemaVersion: "1.0",
		},
	}

	return companyInfo, nil
}
