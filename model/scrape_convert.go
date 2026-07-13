// scrape_convert.go — DTO → model direct converters for the scrape path.
//
// These converters replace the former scrape DTO → emit.Map*DTO (ampy-proto)
// → fromProto* (model) pipeline with a single hop: scrape DTO → model.
// The key naming, Source strings, currency handling and time-window semantics
// are aligned with the prior emit.Map*DTO output so tests/data_correctness_test.go
// behavior is preserved. All monetary values land as float64 on FundamentalsLine
// (model's external surface), decoded from the scrape ScaledDecimal via
// FromScaledDecimal.
//
// MIC is passed in by the caller (facade.Client.Scrape* already infers it via
// inferMICForSymbol); the prior emit layer's normalizeMIC is not reproduced
// here because inference already yields the canonical MIC.
package model

import (
	"strconv"
	"strings"
	"time"
)

// lineFromScaled builds a FundamentalsLine from a scrape *Scaled value. A nil
// pointer yields a zero-valued line (caller decides whether to append it).
// Currency is upper-cased and validated to 3 chars; otherwise blank.
func lineFromScaled(s *Scaled, currency string, ps, pe time.Time, key string) FundamentalsLine {
	if s == nil {
		return FundamentalsLine{Key: key, PeriodStart: ps, PeriodEnd: pe}
	}
	return FundamentalsLine{
		Key:          key,
		Value:        FromScaledDecimal(ScaledDecimal{Scaled: s.Scaled, Scale: s.Scale}),
		CurrencyCode: normalizeCurrency(currency),
		PeriodStart:  ps,
		PeriodEnd:    pe,
	}
}

// lineFromFloat builds a FundamentalsLine directly from a *float64. The prior
// emit layer scaled these (×10^scale into ScaledDecimal) then decoded back to
// float64 — net effect identical to using the raw float, so we skip the
// scaling round-trip.
func lineFromFloat(f *float64, currency string, ps, pe time.Time, key string) FundamentalsLine {
	if f == nil {
		return FundamentalsLine{Key: key, PeriodStart: ps, PeriodEnd: pe}
	}
	return FundamentalsLine{
		Key:          key,
		Value:        *f,
		CurrencyCode: normalizeCurrency(currency),
		PeriodStart:  ps,
		PeriodEnd:    pe,
	}
}

// lineFromInt builds a FundamentalsLine from a whole-number *int64 (counts,
// share counts, analyst counts). Currency is blank for unitless metrics.
func lineFromInt(n *int64, ps, pe time.Time, key string) FundamentalsLine {
	if n == nil {
		return FundamentalsLine{Key: key, PeriodStart: ps, PeriodEnd: pe}
	}
	return FundamentalsLine{
		Key:         key,
		Value:       float64(*n),
		PeriodStart: ps,
		PeriodEnd:   pe,
	}
}

// lineFromIntPtr is the *int variant of lineFromInt for DTO fields typed as
// *int (analyst counts, EPS revision counts). Same semantics.
func lineFromIntPtr(n *int, ps, pe time.Time, key string) FundamentalsLine {
	if n == nil {
		return FundamentalsLine{Key: key, PeriodStart: ps, PeriodEnd: pe}
	}
	return FundamentalsLine{
		Key:         key,
		Value:       float64(*n),
		PeriodStart: ps,
		PeriodEnd:   pe,
	}
}

// normalizeCurrency upper-cases and requires a 3-char ISO-4217 code.
func normalizeCurrency(currency string) string {
	cur := strings.ToUpper(strings.TrimSpace(currency))
	if len(cur) != 3 {
		return ""
	}
	return cur
}

// pointInTimeBounds returns the [start, start+24h) window for point-in-time
// data (key statistics, analysis, analyst insights) — AsOf day in UTC.
func pointInTimeBounds(asOf time.Time) (time.Time, time.Time) {
	start := time.Date(asOf.Year(), asOf.Month(), asOf.Day(), 0, 0, 0, 0, time.UTC)
	return start, start.Add(24 * time.Hour)
}

// quarterBounds returns the [quarterStart, quarterEnd) window for period-based
// financials (income / balance-sheet / cash-flow), aligned with emit's
// extractCurrentPeriodLines: quarterStart = first day of the AsOf quarter,
// quarterEnd = last day of that quarter.
func quarterBounds(asOf time.Time) (time.Time, time.Time) {
	qm := ((asOf.Month()-1)/3)*3 + 1
	qs := time.Date(asOf.Year(), qm, 1, 0, 0, 0, 0, time.UTC)
	return qs, qs.AddDate(0, 3, -1)
}

// parseRevenueEstimateString parses revenue estimate strings like "187.14B",
// "1.2T", "500M", "750K" into a float64. Strips $ prefix and commas. Returns
// ok=false for blank / "--" / "N/A" or unparseable input. Mirrors emit's
// parseRevenueEstimateString but returns float64 instead of *Scaled.
func parseRevenueEstimateString(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" || s == "N/A" {
		return 0, false
	}
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimPrefix(s, "$")

	multiplier := 1.0
	switch {
	case strings.HasSuffix(s, "B"):
		multiplier = 1e9
		s = strings.TrimSuffix(s, "B")
	case strings.HasSuffix(s, "T"):
		multiplier = 1e12
		s = strings.TrimSuffix(s, "T")
	case strings.HasSuffix(s, "M"):
		multiplier = 1e6
		s = strings.TrimSuffix(s, "M")
	case strings.HasSuffix(s, "K"):
		multiplier = 1e3
		s = strings.TrimSuffix(s, "K")
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return val * multiplier, true
}

// parseGrowthPercent parses a percentage string like "12.5%" into a float64
// (the numeric portion). Returns ok=false if not parseable.
func parseGrowthPercent(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasSuffix(s, "%") {
		return 0, false
	}
	val, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
	if err != nil {
		return 0, false
	}
	return val, true
}

// ScrapeFinancialsToSnapshot converts a ComprehensiveFinancialsDTO (income
// statement current period) into a *FundamentalsSnapshot. Key set aligns
// with emit's extractCurrentPeriodLines.
func ScrapeFinancialsToSnapshot(dto *ComprehensiveFinancialsDTO, mic string) *FundamentalsSnapshot {
	if dto == nil {
		return nil
	}
	ps, pe := quarterBounds(dto.AsOf)
	cur := dto.Currency

	var lines []FundamentalsLine
	add := func(key string, s *Scaled) {
		if s == nil {
			return
		}
		lines = append(lines, lineFromScaled(s, cur, ps, pe, key))
	}
	addInt := func(key string, n *int64) {
		if n == nil {
			return
		}
		lines = append(lines, lineFromInt(n, ps, pe, key))
	}

	add("total_revenue", dto.Current.TotalRevenue)
	add("cost_of_revenue", dto.Current.CostOfRevenue)
	add("gross_profit", dto.Current.GrossProfit)
	add("operating_expense", dto.Current.OperatingExpense)
	add("operating_income", dto.Current.OperatingIncome)
	add("net_non_operating_interest_income_expense", dto.Current.NetNonOperatingInterestIncomeExpense)
	add("other_income_expense", dto.Current.OtherIncomeExpense)
	add("pretax_income", dto.Current.PretaxIncome)
	add("tax_provision", dto.Current.TaxProvision)
	add("net_income", dto.Current.NetIncomeCommonStockholders)
	add("eps_basic", dto.Current.BasicEPS)
	add("eps_diluted", dto.Current.DilutedEPS)
	addInt("shares_outstanding_basic", dto.Current.BasicAverageShares)
	addInt("shares_outstanding_diluted", dto.Current.DilutedAverageShares)
	add("total_expenses", dto.Current.TotalExpenses)
	add("normalized_income", dto.Current.NormalizedIncome)
	add("ebit", dto.Current.EBIT)
	add("ebitda", dto.Current.EBITDA)
	add("reconciled_cost_of_revenue", dto.Current.ReconciledCostOfRevenue)
	add("reconciled_depreciation", dto.Current.ReconciledDepreciation)
	add("normalized_ebitda", dto.Current.NormalizedEBITDA)

	return &FundamentalsSnapshot{
		Symbol: dto.Symbol,
		MIC:    mic,
		Source: "yfinance/scrape/comprehensive-financials",
		AsOf:   dto.AsOf,
		Lines:  lines,
	}
}

// ScrapeBalanceSheetToSnapshot converts a ComprehensiveFinancialsDTO (balance
// sheet current period) into a *FundamentalsSnapshot.
func ScrapeBalanceSheetToSnapshot(dto *ComprehensiveFinancialsDTO, mic string) *FundamentalsSnapshot {
	if dto == nil {
		return nil
	}
	ps, pe := quarterBounds(dto.AsOf)
	cur := dto.Currency

	var lines []FundamentalsLine
	add := func(key string, s *Scaled) {
		if s == nil {
			return
		}
		lines = append(lines, lineFromScaled(s, cur, ps, pe, key))
	}

	add("total_assets", dto.Current.TotalAssets)
	add("total_debt", dto.Current.TotalDebt)
	add("shareholders_equity", dto.Current.CommonStockEquity)
	add("working_capital", dto.Current.WorkingCapital)
	add("tangible_book_value", dto.Current.TangibleBookValue)

	return &FundamentalsSnapshot{
		Symbol: dto.Symbol,
		MIC:    mic,
		Source: "yfinance/scrape/balance-sheet",
		AsOf:   dto.AsOf,
		Lines:  lines,
	}
}

// ScrapeCashFlowToSnapshot converts a ComprehensiveFinancialsDTO (cash flow
// current period) into a *FundamentalsSnapshot.
func ScrapeCashFlowToSnapshot(dto *ComprehensiveFinancialsDTO, mic string) *FundamentalsSnapshot {
	if dto == nil {
		return nil
	}
	ps, pe := quarterBounds(dto.AsOf)
	cur := dto.Currency

	var lines []FundamentalsLine
	add := func(key string, s *Scaled) {
		if s == nil {
			return
		}
		lines = append(lines, lineFromScaled(s, cur, ps, pe, key))
	}

	add("operating_cash_flow", dto.Current.OperatingCashFlow)
	add("investing_cash_flow", dto.Current.InvestingCashFlow)
	add("financing_cash_flow", dto.Current.FinancingCashFlow)
	add("free_cash_flow", dto.Current.FreeCashFlow)
	add("capital_expenditure", dto.Current.CapitalExpenditure)

	return &FundamentalsSnapshot{
		Symbol: dto.Symbol,
		MIC:    mic,
		Source: "yfinance/scrape/cash-flow",
		AsOf:   dto.AsOf,
		Lines:  lines,
	}
}

// ScrapeKeyStatisticsToSnapshot converts a ComprehensiveKeyStatisticsDTO into
// a *FundamentalsSnapshot. Valuation ratios carry no currency; monetary
// metrics carry the DTO currency; share count is a unitless int.
func ScrapeKeyStatisticsToSnapshot(dto *ComprehensiveKeyStatisticsDTO, mic string) *FundamentalsSnapshot {
	if dto == nil {
		return nil
	}
	ps, pe := pointInTimeBounds(dto.AsOf)
	cur := dto.Currency

	var lines []FundamentalsLine
	add := func(key string, s *Scaled, currency string) {
		if s == nil {
			return
		}
		lines = append(lines, lineFromScaled(s, currency, ps, pe, key))
	}

	// Market valuation metrics (monetary — carry currency)
	add("market_cap", dto.Current.MarketCap, cur)
	add("enterprise_value", dto.Current.EnterpriseValue, cur)

	// Valuation ratios (unitless)
	add("pe_ratio_trailing", dto.Current.TrailingPE, "")
	add("pe_ratio_forward", dto.Current.ForwardPE, "")
	add("peg_ratio", dto.Current.PEGRatio, "")
	add("price_to_sales", dto.Current.PriceSales, "")
	add("price_to_book", dto.Current.PriceBook, "")
	add("ev_to_revenue", dto.Current.EnterpriseValueRevenue, "")
	add("ev_to_ebitda", dto.Current.EnterpriseValueEBITDA, "")

	// Additional metrics
	add("beta", dto.Additional.Beta, "")
	if dto.Additional.SharesOutstanding != nil {
		// Shares are whole numbers — wrap as a Scale=0 ScaledDecimal.
		shares := &Scaled{Scaled: *dto.Additional.SharesOutstanding, Scale: 0}
		add("shares_outstanding", shares, "")
	}
	add("profit_margin", dto.Additional.ProfitMargin, "")
	add("operating_margin", dto.Additional.OperatingMargin, "")
	add("return_on_assets", dto.Additional.ReturnOnAssets, "")
	add("return_on_equity", dto.Additional.ReturnOnEquity, "")

	return &FundamentalsSnapshot{
		Symbol: dto.Symbol,
		MIC:    mic,
		Source: "yfinance/scrape/key-statistics",
		AsOf:   dto.AsOf,
		Lines:  lines,
	}
}

// ScrapeAnalysisToSnapshot converts a ComprehensiveAnalysisDTO (earnings /
// revenue estimates, EPS trends & revisions) into a *FundamentalsSnapshot.
// EPS estimates (*float64) land as raw float64 (the prior emit scaling
// round-trip was identity for float64).
func ScrapeAnalysisToSnapshot(dto *ComprehensiveAnalysisDTO, mic string) *FundamentalsSnapshot {
	if dto == nil {
		return nil
	}
	ps, pe := pointInTimeBounds(dto.AsOf)

	var lines []FundamentalsLine
	addFloat := func(key string, f *float64, currency string) {
		if f == nil {
			return
		}
		lines = append(lines, lineFromFloat(f, currency, ps, pe, key))
	}
	addIntPtr := func(key string, n *int) {
		if n == nil {
			return
		}
		lines = append(lines, lineFromIntPtr(n, ps, pe, key))
	}

	// Earnings estimates (EPS)
	eeCur := dto.EarningsEstimate.Currency
	addFloat("eps_estimate_current_quarter", dto.EarningsEstimate.CurrentQtr.AvgEstimate, eeCur)
	addFloat("eps_estimate_next_quarter", dto.EarningsEstimate.NextQtr.AvgEstimate, eeCur)
	addFloat("eps_estimate_current_year", dto.EarningsEstimate.CurrentYear.AvgEstimate, eeCur)
	addFloat("eps_estimate_next_year", dto.EarningsEstimate.NextYear.AvgEstimate, eeCur)

	// Analyst counts (unitless)
	addIntPtr("analyst_count_current_quarter", dto.EarningsEstimate.CurrentQtr.NoOfAnalysts)

	// Earnings history — most recent actual EPS
	if len(dto.EarningsHistory.Data) > 0 {
		recent := dto.EarningsHistory.Data[0]
		addFloat("eps_actual_recent", recent.EPSActual, dto.EarningsHistory.Currency)
	}

	// EPS trends
	etCur := dto.EPSTrend.Currency
	addFloat("eps_trend_current_quarter", dto.EPSTrend.CurrentQtr.CurrentEstimate, etCur)
	addFloat("eps_trend_current_quarter_7d_ago", dto.EPSTrend.CurrentQtr.Days7Ago, etCur)
	addFloat("eps_trend_current_quarter_30d_ago", dto.EPSTrend.CurrentQtr.Days30Ago, etCur)
	addFloat("eps_trend_current_year", dto.EPSTrend.CurrentYear.CurrentEstimate, etCur)
	addFloat("eps_trend_next_year", dto.EPSTrend.NextYear.CurrentEstimate, etCur)

	// EPS revisions (Up/Down in last 7/30 days — whole numbers)
	addIntPtr("eps_revisions_up_7d_current_quarter", dto.EPSRevisions.CurrentQtr.UpLast7Days)
	addIntPtr("eps_revisions_down_7d_current_quarter", dto.EPSRevisions.CurrentQtr.DownLast7Days)
	addIntPtr("eps_revisions_up_30d_current_quarter", dto.EPSRevisions.CurrentQtr.UpLast30Days)
	addIntPtr("eps_revisions_down_30d_current_quarter", dto.EPSRevisions.CurrentQtr.DownLast30Days)

	// Revenue estimates (string like "187.14B")
	revCur := dto.RevenueEstimate.Currency
	if v, ok := parseRevenueEstimateString(ptrString(dto.RevenueEstimate.CurrentQtr.AvgEstimate)); ok {
		lines = append(lines, FundamentalsLine{
			Key: "revenue_estimate_current_quarter", Value: v,
			CurrencyCode: normalizeCurrency(revCur), PeriodStart: ps, PeriodEnd: pe,
		})
	}
	if v, ok := parseRevenueEstimateString(ptrString(dto.RevenueEstimate.CurrentYear.AvgEstimate)); ok {
		lines = append(lines, FundamentalsLine{
			Key: "revenue_estimate_current_year", Value: v,
			CurrencyCode: normalizeCurrency(revCur), PeriodStart: ps, PeriodEnd: pe,
		})
	}

	// Growth estimates (percentage string)
	if dto.GrowthEstimate.CurrentYear != nil {
		if v, ok := parseGrowthPercent(*dto.GrowthEstimate.CurrentYear); ok {
			lines = append(lines, FundamentalsLine{
				Key: "growth_estimate_current_year", Value: v,
				PeriodStart: ps, PeriodEnd: pe,
			})
		}
	}

	return &FundamentalsSnapshot{
		Symbol: dto.Symbol,
		MIC:    mic,
		Source: "yfinance/scrape/analysis",
		AsOf:   dto.AsOf,
		Lines:  lines,
	}
}

// ScrapeAnalystInsightsToSnapshot converts an AnalystInsightsDTO (price
// targets + recommendation) into a *FundamentalsSnapshot. Price targets are
// assumed USD (the DTO carries no currency); counts/scores are unitless.
func ScrapeAnalystInsightsToSnapshot(dto *AnalystInsightsDTO, mic string) *FundamentalsSnapshot {
	if dto == nil {
		return nil
	}
	ps, pe := pointInTimeBounds(dto.AsOf)

	var lines []FundamentalsLine
	addPrice := func(key string, f *float64) {
		if f == nil {
			return
		}
		lines = append(lines, lineFromFloat(f, "USD", ps, pe, key))
	}

	addPrice("current_price", dto.CurrentPrice)
	addPrice("target_price_mean", dto.TargetMeanPrice)
	addPrice("target_price_median", dto.TargetMedianPrice)
	addPrice("target_price_high", dto.TargetHighPrice)
	addPrice("target_price_low", dto.TargetLowPrice)

	if dto.NumberOfAnalysts != nil {
		lines = append(lines, lineFromIntPtr(dto.NumberOfAnalysts, ps, pe, "analyst_count"))
	}

	if dto.RecommendationMean != nil {
		lines = append(lines, lineFromFloat(dto.RecommendationMean, "", ps, pe, "recommendation_score"))
	}

	// Upside potential = (target_mean - current) / current * 100
	if dto.CurrentPrice != nil && dto.TargetMeanPrice != nil && *dto.CurrentPrice != 0 {
		upside := ((*dto.TargetMeanPrice - *dto.CurrentPrice) / *dto.CurrentPrice) * 100
		lines = append(lines, FundamentalsLine{
			Key: "upside_potential_percent", Value: upside,
			PeriodStart: ps, PeriodEnd: pe,
		})
	}

	return &FundamentalsSnapshot{
		Symbol: dto.Symbol,
		MIC:    mic,
		Source: "yfinance/scrape/analyst-insights",
		AsOf:   dto.AsOf,
		Lines:  lines,
	}
}

// ScrapeNewsToItems converts a slice of ScrapeNewsItem into model.NewsItem.
// Minimal transform: trims title/source, drops empty-title/empty-URL items,
// dereferences PublishedAt (*time.Time → time.Time, nil → zero). Tracking
// param cleanup / ticker validation that the prior emit layer did are not
// reproduced here — model stays free of net/url dependencies.
func ScrapeNewsToItems(items []ScrapeNewsItem) []NewsItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]NewsItem, 0, len(items))
	for _, it := range items {
		title := strings.TrimSpace(it.Title)
		url := strings.TrimSpace(it.URL)
		if title == "" || url == "" {
			continue
		}
		row := NewsItem{
			Title:   title,
			URL:     url,
			Source:  strings.TrimSpace(it.Source),
			Summary: "",
			Symbols: it.RelatedTickers,
		}
		if it.PublishedAt != nil {
			row.PublishedAt = it.PublishedAt.UTC()
		}
		out = append(out, row)
	}
	return out
}

// ptrString dereferences a *string, returning "" for nil.
func ptrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
